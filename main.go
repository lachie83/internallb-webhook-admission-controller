/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Runtime binary flags
type options struct {
	ServiceAnnotationKey   string
	ServiceAnnotationValue string
	PortNumber             string
}

var (
	// Options runtime binary flags
	Options options
)

// only allow pods to pull images from specific registry.
func admitServices(ar v1beta1.AdmissionReview, o *options) *v1beta1.AdmissionResponse {
	var reviewStatus = &v1beta1.AdmissionResponse{
		Allowed: true,
	}

	// The externalAdmissionHookConfiguration registered via selfRegistration
	// asks the kube-apiserver only sends admission request regarding services.
	serviceResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	if ar.Request.Resource != serviceResource {
		glog.Errorf("expect resource to be %s", serviceResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	service := v1.Service{}
	if err := json.Unmarshal(raw, &service); err != nil {
		glog.Error(err)
		return nil
	}

	if service.Spec.Type == "LoadBalancer" {
		validateLB(reviewStatus, service, o)
	}

	return reviewStatus
}

func validateLB(r *v1beta1.AdmissionResponse, s v1.Service, o *options) {
	r.Allowed = false
	r.Result = &metav1.Status{
		Reason: "the service annotations do not contain required key and value",
	}

	for k, v := range s.ObjectMeta.Annotations {
		if k == o.ServiceAnnotationKey && v == o.ServiceAnnotationValue {
			r.Allowed = true
			r.Result = nil
		}
	}
}

func mutateServices(ar v1beta1.AdmissionReview, o *options) *v1beta1.AdmissionResponse {
	var reviewResponse = &v1beta1.AdmissionResponse{
		Allowed: true,
	}

	serviceResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	if ar.Request.Resource != serviceResource {
		glog.Errorf("expect resource to be %s", serviceResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	service := v1.Service{}
	if err := json.Unmarshal(raw, &service); err != nil {
		glog.Error(err)
		return nil
	}

	addServiceAnnotationPatch := fmt.Sprintf(`[
		 {"op":"add","path":"/metadata/annotations","value":{"%s":"%s"}}
	]`, o.ServiceAnnotationKey, o.ServiceAnnotationValue)

	if service.Spec.Type == "LoadBalancer" {
		glog.V(2).Infof("patching service type LoadBalancer name: %v", service.ObjectMeta.Name)
		reviewResponse.Patch = []byte(addServiceAnnotationPatch)
		pt := v1beta1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
	}

	return reviewResponse
}

type admitFunc func(v1beta1.AdmissionReview, *options) *v1beta1.AdmissionResponse

func serve(w http.ResponseWriter, r *http.Request, o *options, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(body, &ar); err != nil {
		glog.Error(err)
		reviewResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		reviewResponse = admit(ar, o)
	}

	response := v1beta1.AdmissionReview{
		Response: reviewResponse,
	}

	resp, err := json.Marshal(response)
	if err != nil {
		glog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func serveServices(w http.ResponseWriter, r *http.Request) {
	serve(w, r, &Options, admitServices)
}

func serveMutateServices(w http.ResponseWriter, r *http.Request) {
	serve(w, r, &Options, mutateServices)
}

func serveHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func main() {
	certKey := certKey{}
	flag.StringVar(&Options.PortNumber, "port", "8443", "webserver port")
	flag.StringVar(&certKey.PairName, "keypairname", "tls", "certificate and key pair name")
	flag.StringVar(&certKey.CertDirectory, "certdir", "/var/run/internallb-webhook-admission-controller", "certificate and key directory")
	flag.StringVar(&Options.ServiceAnnotationKey, "svcannotationkey", "service.beta.kubernetes.io/azure-load-balancer-internal", "service annotation key to match or mutate")
	flag.StringVar(&Options.ServiceAnnotationValue, "svcannotationvalue", "true", "service annotation value to match or mutate")
	flag.Parse()

	http.HandleFunc("/services", serveServices)
	http.HandleFunc("/mutating-services", serveMutateServices)
	http.HandleFunc("/healthz", serveHealthz)
	clientset := getClient()
	server := &http.Server{
		Addr:      fmt.Sprintf(":%s", Options.PortNumber),
		TLSConfig: configTLS(clientset, &certKey),
	}

	glog.V(2).Infof("starting webserver on port %s", Options.PortNumber)
	glog.V(2).Infof("service annotation to match/mutate: %s: %s", Options.ServiceAnnotationKey, Options.ServiceAnnotationValue)

	if err := server.ListenAndServeTLS("", ""); err != nil {
		glog.Fatal(err)
	}

}
