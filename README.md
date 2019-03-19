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
