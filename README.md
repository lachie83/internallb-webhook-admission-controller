
# Kubernetes Internal LoadBalancer Admission Webhook

This Kubernetes Admission controller only admits the creation of v1/service resources containing the correct cloud provider annotations to create Internal LoadBalancers.

See [upstream k8s docs](https://kubernetes.io/docs/concepts/services-networking/service/#internal-load-balancer) for details on these annotations

## Project State

Experimental

## Attribution

This projects uses the upstream examples found in the following repos:
* https://github.com/caesarxuchao/example-webhook-admission-controller
* https://github.com/kubernetes/kubernetes/tree/release-1.9/test/images/webhook

Massive thanks for all the work that went into crafting reusable examples.

## Supported Kubernetes versions

* 1.9

## Supported Clouds

* Azure

## Prerequisites
Please enable the admission webhook feature
[doc](https://kubernetes.io/docs/admin/extensible-admission-controllers/#enable-external-admission-webhooks).

## Build

```bash
make build
```

## Deploy

This project comes with a Helm chart to simplify deployment. The chart will
generate the required certificates and keys, put them into secrets, and create
a pod that mounts the secrets for the admission webhook to access.

```bash
helm install ./helm
```

## Explanation on the CAs/Certs/Keys

Taken from upstream https://github.com/caesarxuchao/example-webhook-admission-controller

The apiserver initiates a tls connection with the webhook, so the apiserver is
the tls client, and the webhook is the tls server.

The webhook proves its identity by the `serverCert` in the certs.go. The server
cert is signed by the CA in certs.go. To let the apiserver trust the `caCert`,
the webhook registers itself with the apiserver via the
`admissionregistration/v1alpha1/externalAdmissionHook` API, with
`clientConfig.caBundle=caCert`.

For maximum protection, this example webhook requires and verifies the client
(i.e., the apiserver in this case) cert. The cert presented by the apiserver is
signed by a client CA, whose cert is stored in the configmap
`extension-apiserver-authentication` in the `kube-system` namespace. See the
`getAPIServerCert` function for more information. Usually you don't need to
worry about setting up this CA cert. It's taken care of when the cluster is
created. You can disable the client cert verification by setting the
`tls.Config.ClientAuth` to `tls.NoClientCert` in `config.go`.
