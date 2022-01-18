module cdap.io/cdap-operator

go 1.12

require (
	github.com/go-logr/logr v0.1.0
	github.com/nsf/jsondiff v0.0.0-20190712045011-8443391ee9b6
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/controller-reconciler v0.0.0
	sigs.k8s.io/controller-runtime v0.4.0
)

replace sigs.k8s.io/controller-reconciler => ./vendor/sigs.k8s.io/controller-reconciler
