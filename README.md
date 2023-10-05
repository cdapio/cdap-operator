# Kubernetes operator for [CDAP](http://cdap.io)

## Project Status

*Alpha*

The CDAP Operator is still under active development and has not been extensively tested in production environment. Backward compatibility of the APIs is not guaranteed for alpha releases.

## Prerequisites
* Version >= 1.9 of Kubernetes.
* Version >= 6.0.0 of CDAP

## Quick Start

### Build and Run Locally

You can checkout the CDAP Operator source code, build and run locally. To build the CDAP Operator, you need to setup your environment for the [Go](https://golang.org/doc/install) language. Also, you should have a Kubernetes cluster

1. Checkout CDAP Operator source
   ```
   mkdir -p $GOPATH/src/cdap.io
   cd $GOPATH/src/cdap.io
   git clone https://github.com/cdapio/cdap-operator.git
   cd cdap-operator
   ```
1. Generates and install the CRDs
   ```
   make install
   ```
1. Compiles and run the CDAP Operator
   ```
   make run
   ```
1. Deploy CDAP CRD to the cluster
   ```
   kubectl apply -k config/crd
   ```
1. Edit the sample CDAP CR and deploy to the cluster
   ```
   kubectl apply -f config/samples/cdap_v1alpha1_cdapmaster.yaml
   ```

### Build Controller Docker Image and Deploy in Kubernetes

You can also build a docker image containing the CDAP controller and deploy it to Kubernetes.

1. Build the docker image
   ```
   IMG=cdap-controller:latest make docker-build
   ```
   You can change the target image name and tag by setting the `IMG` environment variable.
1. Push the docker image
   ```
   IMG=cdap-controller:latest make docker-push
   ```
1. Deploy CDAP CRD and RBAC to the cluster
   ```
   make deploy
   ```

### Using CDAP operator to manage CDAP instances in Kubernetes

A step by step guide of running CDAP in Kubernetes using CDAP operator can be found in the [blog post](https://link.medium.com/hpPbiUYT9X).

### Using the Admission Controller

The CDAP operator can be configured to optionally run a webhook server for a [mutating admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/). The mutating admission controller allows the operator to change the following fields in CDAP pods:
1. Add [init containers](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/)
1. Add [Node Selectors](https://kubernetes.io/docs/tasks/configure-pod-container/assign-pods-nodes/)
1. Add [tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/)

These mutations can be defined using the `MutationConfigs` field in CDAPMaster.

#### Prerequisites

Kubernetes requires that the webhook server uses TLS to authenticate with the kube API server. For this you will need to ensure the TLS certificates are present in the `/tmp/k8s-webhook-server/serving-certs` directory in the `cdap-controller` pod. To simplify the management of TLS certificates, you can use [cert-manager](https://github.com/cert-manager/cert-manager). The following steps assume you are in the root directory of the Git repository and have already deployed the CDAP operator stateful set.
1. [Deploy cert-manager](https://cert-manager.io/docs/installation/#default-static-install) in the cluster.
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml
```
You should see 3 pods running for cert-manager.
```bash
kubectl get pods -n cert-manager
NAME                                       READY   STATUS    RESTARTS       AGE
cert-manager-655c4cf99d-rbzwr              1/1     Running   0              2m
cert-manager-cainjector-845856c584-csbsw   1/1     Running   0              2m
cert-manager-webhook-57876b9fd-68vgc       1/1     Running   0              2m
```
2. Deploy a kubernetes service for the webhook server.
```bash
# set the namespace in which CDAPMaster is deployed.
export CDAP_NAMESPACE=default
sed -e 's@{CDAP_NAMESPACE}@'"$CDAP_NAMESPACE"'@g' <"./webhooks/templates/webhook-service.yaml" | kubectl apply -f -
```
3. Deploy the cert-manager self-signed issuer.
```bash
sed -e 's@{CDAP_NAMESPACE}@'"$CDAP_NAMESPACE"'@g' <"./webhooks/templates/issuer.yaml" | kubectl apply -f -
```
4. Deploy the Certificate resource.
```bash
sed -e 's@{CDAP_NAMESPACE}@'"$CDAP_NAMESPACE"'@g' <"./webhooks/templates/certificate.yaml" | kubectl apply -f -
```
Wait for the certificate to be ready.
```bash
kubectl get Certificates
NAME                READY   SECRET                     AGE
cdap-webhook-cert   True    cdap-webhook-server-cert   1d
```
5. Add the following fields in the CDAP operator stateful set spec:
```yaml
# Filename: cdap-controller.yaml
spec:
  containers:
  - command:
    - /manager
    args: ["--enable-webhook", "true"]
...
    volumeMounts:
    - mountPath: /tmp/k8s-webhook-server/serving-certs
      name: cert
      readOnly: true
...
  volumes:
  - name: cert
    secret:
      defaultMode: 420
      secretName: cdap-webhook-server-cert

```
6. Deploy the mutating webhook resource:
```bash
sed -e 's@{CDAP_NAMESPACE}@'"$CDAP_NAMESPACE"'@g' <"./webhooks/templates/webhook.yaml" | kubectl apply -f -
```
The webhook is now configured and it will intercept requests to create new pods made by CDAP.

#### Example use case: Isolate pods that execute user code in Google Kubernetes Engine.

Assuming task workers are enabled, the pods that execute user code in CDAP are task workers and preview runners. Let us call these pods as "worker pods". To isolate these worker pods in a dedicated node pool with the help of the admission controller, you follow these steps:
1. Create a node pool for running only worker pods.
```bash
gcloud container node-pools create worker-pool \
    --cluster cdap-cluster --project my-gcp-projet --location us-east1
```
2. Add a taint to the new node pool. This will prevent pods from being scheduled on the node pool unless they specify the corresponding toleration.
```bash
gcloud beta container node-pools update worker-pool  \
--node-taints="worker-pods-only=true:NoExecute" \
--cluster cdap-cluster --project my-gcp-projet --location us-east1
```
3. Add the following configuration to the CDAPMaster:
```yaml
# Filename: cdapmaster.yaml
spec:
...
  mutationConfigs:
  - labelSelector:
      matchExpressions:
      - {key: cdap.twill.app, operator: In, values: [task.worker, preview.runner]}
    podMutations:
      nodeSelectors:
        cloud.google.com/gke-nodepool: worker-pool
      tolerations:
      - effect: NoExecute
        key: worker-pods-only
        operator: Equal
        tolerationSeconds: 3600
        value: "true"
```
Now whenever CDAP launches preview runner of task worker pods, the admission controller will mutate the pod specifications before they are deployed to ensure the pods get scheduled only on the node pool "worker-pool".
### Running Unit Tests

1. Install [kubebuilder](https://book-v1.book.kubebuilder.io/quick_start.html).

2. Install [setup-envtest](https://github.com/kubernetes-sigs/controller-runtime/tree/master/tools/setup-envtest#envtest-binaries-manager) by running:
```
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
```

3. After installing `setup-envtest`, use it to download envtest 1.19.x for kubebuilder and to set your KUBEBUILDER_ASSETS environment variable:
```bash
# Downloads envtest v1.19.x and writes the export statement to a temporary file
$(go env GOPATH)/bin/setup-envtest use -p env 1.19.x > /tmp/setup_envtest.sh
# Sets the KUBEBUILDER_ASSETS environment variable
source /tmp/setup_envtest.sh
# Deletes the temporary file
rm /tmp/setup_envtest.sh
```

4. Run `make test`
