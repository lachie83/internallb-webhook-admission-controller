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
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type v1Service struct {
	Spec              v1.ServiceSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status            v1.ServiceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
	metav1.ObjectMeta `json:",inline"`
	metav1.TypeMeta   `json:",inline"`
}

// only allow pods to pull images from specific registry.
func admit(data []byte) *v1beta1.AdmissionResponse {
	var reviewStatus = &v1beta1.AdmissionResponse{
		Allowed: true,
	}

	ar := v1beta1.AdmissionReview{}
	if err := json.Unmarshal(data, &ar); err != nil {
		glog.Error(err)
		return nil
	}
	// The externalAdmissionHookConfiguration registered via selfRegistration
	// asks the kube-apiserver only sends admission request regarding services.
	serviceResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	if ar.Request.Resource != serviceResource {
		glog.Errorf("expect resource to be %s", serviceResource)
		return nil
	}

	raw := ar.Request.Object.Raw
	service := v1Service{}
	if err := json.Unmarshal(raw, &service); err != nil {
		glog.Error(err)
		return nil
	}

	if service.Spec.Type == "LoadBalancer" {
		admitLB(reviewStatus, service)
	}

	return reviewStatus
}

func admitLB(r *v1beta1.AdmissionResponse, s v1Service) {
	r.Allowed = false
	r.Result = &metav1.Status{
		Reason: "the service annotations do not contain required key and value",
	}

	for k, v := range s.ObjectMeta.Annotations {
		if k == "service.beta.kubernetes.io/azure-load-balancer-internal" && v == "true" {
			r.Allowed = true
			r.Result = nil
		}
	}
}

func serve(w http.ResponseWriter, r *http.Request) {
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

	reviewResponse := admit(body)
	ar := v1beta1.AdmissionReview{
		Response: reviewResponse,
	}

	resp, err := json.Marshal(ar)
	if err != nil {
		glog.Error(err)
	}
	if _, err := w.Write(resp); err != nil {
		glog.Error(err)
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/", serve)
	clientset := getClient()
	server := &http.Server{
		Addr:      ":8000",
		TLSConfig: configTLS(clientset),
	}
	go selfRegistration(clientset, caCert)
	server.ListenAndServeTLS("", "")
}
