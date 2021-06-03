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
   kubectl apply -f config/crds
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

### Running Unit Tests

1. Install [kubebuilder](https://book-v1.book.kubebuilder.io/quick_start.html).

2. Run `make test`