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
	"io/ioutil"

	"github.com/golang/glog"
)

var caCert, serverKey, serverCert []byte

func init() {
	var err error

	if caCert, err = ioutil.ReadFile("/secrets/certs/caCert"); err != nil {
		glog.Fatal(err)
	}

	if serverKey, err = ioutil.ReadFile("/secrets/server-key/key"); err != nil {
		glog.Fatal(err)
	}

	if serverCert, err = ioutil.ReadFile("/secrets/certs/serverCert"); err != nil {
		glog.Fatal(err)
	}
}
