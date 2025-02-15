package v1beta3

import (
	"context"
	"fmt"
	k8sCoreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchTypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sWatchV1 "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	tarsCrdV1beta3 "k8s.tars.io/crd/v1beta3"
	tarsMetaTools "k8s.tars.io/meta/tools"
	tarsMetaV1beta3 "k8s.tars.io/meta/v1beta3"
	"tarscontroller/controller"
	"tarscontroller/reconcile"
	"time"
)

type TTreeReconciler struct {
	clients   *controller.Clients
	informers *controller.Informers
	threads   int
	workQueue workqueue.RateLimitingInterface
}

func NewTTreeReconciler(clients *controller.Clients, informers *controller.Informers, threads int) *TTreeReconciler {
	reconciler := &TTreeReconciler{
		clients:   clients,
		informers: informers,
		threads:   threads,
		workQueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ""),
	}
	informers.Register(reconciler)
	return reconciler
}

func (r *TTreeReconciler) processItem() bool {

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
	default:
		//code should not reach here
		utilRuntime.HandleError(fmt.Errorf("should not reach place"))
		return false
	}
}

func (r *TTreeReconciler) EnqueueObj(resourceName string, resourceEvent k8sWatchV1.EventType, resourceObj interface{}) {
	switch resourceObj.(type) {
	case *tarsCrdV1beta3.TServer:
		tserver := resourceObj.(*tarsCrdV1beta3.TServer)
		key := fmt.Sprintf("%s/%s", tserver.Namespace, tserver.Name)
		r.workQueue.Add(key)
	default:
		return
	}
}

func (r *TTreeReconciler) Start(stopCh chan struct{}) {
	for i := 0; i < r.threads; i++ {
		workFun := func() {
			for r.processItem() {
			}
			r.workQueue.ShutDown()
		}
		go wait.Until(workFun, time.Second, stopCh)
	}
}

func (r *TTreeReconciler) reconcile(key string) reconcile.Result {
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
		return reconcile.AllOk
	}

	if tserver.DeletionTimestamp != nil {
		return reconcile.AllOk
	}

	ttree, err := r.informers.TTreeInformer.Lister().TTrees(namespace).Get(tarsMetaV1beta3.FixedTTreeResourceName)
	if err != nil {
		msg := fmt.Sprintf(tarsMetaV1beta3.ResourceGetError, "ttree", namespace, tarsMetaV1beta3.FixedTTreeResourceName, err.Error())
		utilRuntime.HandleError(fmt.Errorf(msg))
		controller.Event(tserver, k8sCoreV1.EventTypeWarning, tarsMetaV1beta3.ResourceGetReason, msg)
		return reconcile.RateLimit
	}

	for i := range ttree.Apps {
		if ttree.Apps[i].Name == tserver.Spec.App {
			return reconcile.AllOk
		}
	}

	newTressApp := &tarsCrdV1beta3.TTreeApp{
		Name:         tserver.Spec.App,
		BusinessRef:  "",
		CreatePerson: "",
		CreateTime:   k8sMetaV1.Now(),
		Mark:         "AddByController",
	}
	jsonPatch := tarsMetaTools.JsonPatch{
		{
			OP:    tarsMetaTools.JsonPatchAdd,
			Path:  "/apps/-",
			Value: newTressApp,
		},
	}

	patchContent, _ := json.Marshal(jsonPatch)
	_, err = r.clients.CrdClient.CrdV1beta3().TTrees(namespace).Patch(context.TODO(), tarsMetaV1beta3.FixedTTreeResourceName, patchTypes.JSONPatchType, patchContent, k8sMetaV1.PatchOptions{})
	if err != nil {
		return reconcile.RateLimit
	}

	return reconcile.AllOk
}
