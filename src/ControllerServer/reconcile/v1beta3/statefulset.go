package v1beta3

import (
	"context"
	"fmt"
	k8sAppsV1 "k8s.io/api/apps/v1"
	k8sCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sWatchV1 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	tarsCrdV1beta3 "k8s.tars.io/crd/v1beta3"
	tarsMetaV1beta3 "k8s.tars.io/meta/v1beta3"
	"tarscontroller/controller"
	"tarscontroller/reconcile"
	"time"
)

type StatefulSetReconciler struct {
	clients   *controller.Clients
	informers *controller.Informers
	threads   int
	workQueue workqueue.RateLimitingInterface
}

func diffVolumeClaimTemplate(currents, targets []k8sCoreV1.PersistentVolumeClaim) (bool, []string) {
	if currents == nil && targets != nil {
		return false, nil
	}

	targetVCS := make(map[string]*k8sCoreV1.PersistentVolumeClaim, len(targets))
	for i := range targets {
		targetVCS[targets[i].Name] = &targets[i]
	}

	var equal = true
	var shouldDelete []string

	for i := range currents {
		c := &currents[i]
		t, ok := targetVCS[c.Name]
		if !ok {
			equal = false
			shouldDelete = append(shouldDelete, c.Name)
			continue
		}

		if equal == true {
			if !equality.Semantic.DeepEqual(c.ObjectMeta, t.ObjectMeta) {
				equal = false
				continue
			}

			if !equality.Semantic.DeepEqual(c.Spec, t.Spec) {
				equal = false
				continue
			}
		}
	}
	return equal, shouldDelete
}

func NewStatefulSetReconciler(clients *controller.Clients, informers *controller.Informers, threads int) *StatefulSetReconciler {
	reconciler := &StatefulSetReconciler{
		clients:   clients,
		informers: informers,
		threads:   threads,
		workQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ""),
	}
	informers.Register(reconciler)
	return reconciler
}

func (r *StatefulSetReconciler) processItem() bool {

	obj, shutdown := r.workQueue.Get()

	if shutdown {
		return false
	}

	defer r.workQueue.Done(obj)

	key, ok := obj.(string)
	if !ok {
		utilRuntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		r.workQueue.Forget(obj)
		return true
	}

	res := r.reconcile(key)

	switch res {
	case reconcile.AllOk:
		r.workQueue.Forget(obj)
		return true
	case reconcile.RateLimit:
		r.workQueue.AddRateLimited(obj)
		return true
	case reconcile.FatalError:
		r.workQueue.ShutDown()
		return false
	case reconcile.AddAfter:
		r.workQueue.AddAfter(obj, time.Second*3)
		return true
	default:
		//code should not reach here
		utilRuntime.HandleError(fmt.Errorf("should not reach place"))
		return false
	}
}

func (r *StatefulSetReconciler) EnqueueObj(resourceName string, resourceEvent k8sWatchV1.EventType, resourceObj interface{}) {
	switch resourceObj.(type) {
	case *tarsCrdV1beta3.TServer:
		tserver := resourceObj.(*tarsCrdV1beta3.TServer)
		key := fmt.Sprintf("%s/%s", tserver.Namespace, tserver.Name)
		r.workQueue.Add(key)
	case *k8sAppsV1.StatefulSet:
		statefulset := resourceObj.(*k8sAppsV1.StatefulSet)
		if statefulset.DeletionTimestamp != nil || resourceEvent == k8sWatchV1.Deleted {
			key := fmt.Sprintf("%s/%s", statefulset.Namespace, statefulset.Name)
			r.workQueue.Add(key)
		}
	}
}

func (r *StatefulSetReconciler) Start(stopCh chan struct{}) {
	for i := 0; i < r.threads; i++ {
		workFun := func() {
			for r.processItem() {
			}
			r.workQueue.ShutDown()
		}
		go wait.Until(workFun, time.Second, stopCh)
	}
}

func (r *StatefulSetReconciler) syncStatefulset(tserver *tarsCrdV1beta3.TServer, statefulSet *k8sAppsV1.StatefulSet, namespace, name string) reconcile.Result {
	statefulSetCopy := statefulSet.DeepCopy()
	syncStatefulSet(tserver, statefulSetCopy)
	statefulSetInterface := r.clients.K8sClient.AppsV1().StatefulSets(namespace)
	if _, err := statefulSetInterface.Update(context.TODO(), statefulSetCopy, k8sMetaV1.UpdateOptions{}); err != nil {
		utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceUpdateError, "statefulset", namespace, name, err.Error()))
		return reconcile.RateLimit
	}
	return reconcile.AllOk
}

func (r *StatefulSetReconciler) recreateStatefulset(tserver *tarsCrdV1beta3.TServer, shouldDeletePVCS []string) reconcile.Result {
	namespace := tserver.Namespace
	name := tserver.Name
	err := r.clients.K8sClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, k8sMetaV1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceDeleteError, "statefulset", namespace, name, err.Error()))
		return reconcile.RateLimit
	}

	if shouldDeletePVCS != nil {
		appRequirement, _ := labels.NewRequirement(tarsMetaV1beta3.TServerAppLabel, selection.DoubleEquals, []string{tserver.Spec.App})
		serverRequirement, _ := labels.NewRequirement(tarsMetaV1beta3.TServerNameLabel, selection.DoubleEquals, []string{tserver.Spec.Server})
		labelSelector := labels.NewSelector().Add(*appRequirement, *serverRequirement)

		localVolumeRequirement, _ := labels.NewRequirement(tarsMetaV1beta3.TLocalVolumeLabel, selection.DoubleEquals, shouldDeletePVCS)
		labelSelector.Add(*localVolumeRequirement)

		err = r.clients.K8sClient.CoreV1().PersistentVolumeClaims(namespace).DeleteCollection(context.TODO(), k8sMetaV1.DeleteOptions{}, k8sMetaV1.ListOptions{
			LabelSelector: labelSelector.String(),
		})

		if err != nil {
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceDeleteCollectionError, "persistentvolumeclaims", labelSelector.String(), err.Error()))
		}
	}
	return reconcile.AddAfter
}

func (r *StatefulSetReconciler) reconcile(key string) reconcile.Result {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilRuntime.HandleError(fmt.Errorf("invalid key: %s", key))
		return reconcile.AllOk
	}

	tserver, err := r.informers.TServerInformer.Lister().TServers(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceGetError, "tserver", namespace, name, err.Error()))
			return reconcile.RateLimit
		}
		err = r.clients.K8sClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, k8sMetaV1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceDeleteError, "statefulset", namespace, name, err.Error()))
			return reconcile.RateLimit
		}
		return reconcile.AllOk
	}

	if tserver.DeletionTimestamp != nil || tserver.Spec.K8S.DaemonSet {
		err = r.clients.K8sClient.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, k8sMetaV1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceDeleteError, "statefulset", namespace, name, err.Error()))
			return reconcile.RateLimit
		}
		return reconcile.AllOk
	}

	statefulSet, err := r.informers.StatefulSetInformer.Lister().StatefulSets(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceGetError, "statefulset", namespace, name, err.Error()))
			return reconcile.RateLimit
		}

		if !tserver.Spec.K8S.DaemonSet {
			statefulSet = buildStatefulset(tserver)
			statefulSetInterface := r.clients.K8sClient.AppsV1().StatefulSets(namespace)
			if _, err = statefulSetInterface.Create(context.TODO(), statefulSet, k8sMetaV1.CreateOptions{}); err != nil && !errors.IsAlreadyExists(err) {
				utilRuntime.HandleError(fmt.Errorf(tarsMetaV1beta3.ResourceCreateError, "statefulset", namespace, name, err.Error()))
				return reconcile.RateLimit
			}
		}
		return reconcile.AllOk
	}

	if statefulSet.DeletionTimestamp != nil {
		return reconcile.AddAfter
	}

	if !k8sMetaV1.IsControlledBy(statefulSet, tserver) {
		// 此处意味着出现了非由 controller 管理的同名 statefulSet, 需要警告和重试
		msg := fmt.Sprintf(tarsMetaV1beta3.ResourceOutControlError, "statefulset", namespace, statefulSet.Name, namespace, name)
		controller.Event(tserver, k8sCoreV1.EventTypeWarning, tarsMetaV1beta3.ResourceOutControlReason, msg)
		return reconcile.RateLimit
	}

	volumeClaimTemplates := buildStatefulsetVolumeClaimTemplates(tserver)
	equal, names := diffVolumeClaimTemplate(statefulSet.Spec.VolumeClaimTemplates, volumeClaimTemplates)
	if !equal {
		return r.recreateStatefulset(tserver, names)
	}

	anyChanged := !EqualTServerAndStatefulSet(tserver, statefulSet)
	if anyChanged {
		return r.syncStatefulset(tserver, statefulSet, namespace, name)
	}
	return reconcile.AllOk
}
