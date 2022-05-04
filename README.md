# Kubernetes operator for [CDAP](http://cdap.io)

## Project Status

*Alpha*

The CDAP Operator is still under active development and has not been extensively tested in production environment. Backward compatibility of the APIs is not guaranteed for alpha releases.

## Prerequisites
* Version >= 1.9 of Kubernetes.
* Version >= 6.0.0 of CDAP

## Quick Start

### Repository Setup

The core of this codebase has contributors both in `LiveRamp` and `cdapio` Github organisations so we have customised the Git setup to enable LiveRamp devs to pull request against an internal production branch called `lr-main` whereas `develop` tracks an upstream branch, namely `cdapio/cdap-operator/tree/develop`. This is enabled through a Github action that runs sync upstream tags and the `remote` branch every night. 

### Local Git Setup

Given the repository setup we reccomend the following local git setup: 

```
origin	git@github.com:LiveRamp/cdap-operator.git (fetch)
origin	git@github.com:LiveRamp/cdap-operator.git (push)
upstream	https://github.com/cdapio/cdap-operator (fetch)
upstream	DISABLE (push)
```

This enables a developer to develop and PR against the LiveRamp "fork" as the `origin` while the `upstream` managed by the Google CDAP team can be pulled in at any time as well. Although not strictly necessary the following can be achieved by cloning this repo and then running

`git remote add upstream https://github.com/cdapio/cdap-operator.git`

`git fetch upstream`

`git remote --push upstream no-pushing`

### CI

Automated CI for this project runs for PR and Branch commits against `lr-main` as the base branch. CI is enabled through Google Cloud Build. Images are currently pushed to `eu.gcr.io/liveramp-eng-data-ops-dev/cdap-controller` , a container registry managed by the DataOps team. Builds off of `lr-main` are pushed with a `latest` tag whereas builds from PRs are pushed with tag `{pr_number}`. It is reccomended for production deploys a purpose built tag be used. Please see `cloudbuild.yaml` and the [cloud build triggers](https://console.cloud.google.com/cloud-build/triggers?project=liveramp-eng-data-ops-dev) config for more details on the build configuration.

### CD

CD for this repo is managed seperately through the `LiveRamp/dop-infra` repository. This is so that we can have both the `cdap-operator` and `cdap` deployments managed through a single helm configuration. They are usually deployed in conjunction. 

In the future we may look to change the CI process to use a LiveRamp build tool like Jenkins but at the moment this does not add any additional value for us as the `cdap-operator` codebase does not depend on any LiveRamp artifacts.

For any PR against `lr-main` you may manually trigger a one off build by typing `/gcbrun` as a comment.

### Build and Run Locally

You can checkout the CDAP Operator source code, build and run locally. To build the CDAP Operator, you need to setup your environment for the [Go](https://golang.org/doc/install) language. Also, you should have a Kubernetes cluster running and accessible in your current `kubectl` context. This could be either a Kubernetes cluster hosted on a cloud platform like GCP GKE or on a local Minikube instance.

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
   
### Build Controller Docker Image and Deploy in Kubernetes Manually

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
