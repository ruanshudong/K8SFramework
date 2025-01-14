package validating

import (
	"encoding/json"
	"fmt"
	k8sAdmissionV1 "k8s.io/api/admission/v1"
	k8sMetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	tarsMetaV1beta2 "k8s.tars.io/meta/v1beta2"
	tarsMetaV1beta3 "k8s.tars.io/meta/v1beta3"
	"net/http"
	"tarscontroller/controller"
	validatingAppsV1 "tarscontroller/webhook/validating/apps/v1"
	validatingCoreV1 "tarscontroller/webhook/validating/core/v1"
	validatingCrdV1Beta2 "tarscontroller/webhook/validating/k8s.tars.io/v1beta2"
	validatingCrdV1Beta3 "tarscontroller/webhook/validating/k8s.tars.io/v1beta3"
)

type Validating struct {
	clients   *controller.Clients
	informers *controller.Informers
}

func New(clients *controller.Clients, informers *controller.Informers) *Validating {
	return &Validating{
		clients:   clients,
		informers: informers,
	}
}

var handlers = map[string]func(*controller.Clients, *controller.Informers, *k8sAdmissionV1.AdmissionReview) error{}

func init() {
	handlers = map[string]func(*controller.Clients, *controller.Informers, *k8sAdmissionV1.AdmissionReview) error{
		"core/v1":                    validatingCoreV1.Handler,
		"/v1":                        validatingCoreV1.Handler,
		"apps/v1":                    validatingAppsV1.Handler,
		tarsMetaV1beta2.GroupVersion: validatingCrdV1Beta2.Handler,
		tarsMetaV1beta3.GroupVersion: validatingCrdV1Beta3.Handler,
	}
}

func (v *Validating) Handle(w http.ResponseWriter, r *http.Request) {
	requestView := &k8sAdmissionV1.AdmissionReview{}

	err := json.NewDecoder(r.Body).Decode(requestView)
	if err != nil {
		return
	}

	gv := fmt.Sprintf("%s/%s", requestView.Request.Kind.Group, requestView.Request.Kind.Version)
	if fun, ok := handlers[gv]; !ok {
		err = fmt.Errorf("unsupported validating %s.%s", gv, requestView.Request.Kind.Kind)
	} else {
		err = fun(v.clients, v.informers, requestView)
	}

	var responseView = &k8sAdmissionV1.AdmissionReview{
		TypeMeta: k8sMetaV1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: &k8sAdmissionV1.AdmissionResponse{
			UID: requestView.Request.UID,
		},
	}
	if err != nil {
		responseView.Response.Allowed = false
		responseView.Response.Result = &k8sMetaV1.Status{
			Status:  "Failure",
			Message: err.Error(),
		}
	} else {
		responseView.Response.Allowed = true
	}
	responseBytes, _ := json.Marshal(responseView)
	_, _ = w.Write(responseBytes)
}
