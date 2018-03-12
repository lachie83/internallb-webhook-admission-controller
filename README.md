
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

There are two types of Webhook Admission controllers in Kubernetes 1.9.
* ValidatingAdmissionWebhook
* MutatingAdmissionWebhook

Enable the relevant Kubeneretes Admission controller by adding to following `--admission-control` and restarting kube-apiserver. See the relevant [docs](https://kubernetes.io/docs/admin/extensible-admission-controllers/#external-admission-webhooks)
```
ValidatingAdmissionWebhook,MutatingAdmissionWebhook
```

Here is an example minikube command to buid a cluster with the Admission Controller flags already present on the API server.
```
minikube start --kubernetes-version v1.9.3 --bootstrapper kubeadm --logtostderr --vm-driver=virtualbox --extra-config=apiserver.admission-control="NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota,ValidatingAdmissionWebhook,MutatingAdmissionWebhook,PodPreset"
```

Once the cluster has been configured you can deploy the admission webhook to using Helm. The default installation configures a `MutatingWebhookConfiguration`.

```
helm install --name admission-webhook charts/internallb-webhook-admission-controller
```

To install a `ValidatingWebhookConfiguration` please use the following command

```
helm install --name admission-webhook charts/internallb-webhook-admission-controller --set admissionRegistration.kind=ValidatingWebhookConfiguration
```

For a full list of configurable values in the helm chart please, run the following command

```
helm inspect charts/internallb-webhook-admission-controller
```
