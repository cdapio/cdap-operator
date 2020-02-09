package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	"fmt"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"strings"
)

type Pair struct {
	first, second interface{}
}

// Creates a int32 pointer for the given value
func int32Ptr(value int32) *int32 {
	return &value
}

func getObjectName(masterName, name string) string {
	return fmt.Sprintf("%s%s-%s", objectNamePrefix, masterName, strings.ToLower(name))
}

func mergeMaps(current, added map[string]string) map[string]string {
	labels := make(reconciler.KVMap)
	labels.Merge(current, added)
	return labels
}

func getCDAPServiceSpec(serviceName ServiceName, master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
	serviceSpecMap := map[ServiceName]func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec{
		serviceAppFabric: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.AppFabric.CDAPServiceSpec
		},
		serviceLogs: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.Logs.CDAPServiceSpec
		},
		serviceMessaging: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.Messaging.CDAPServiceSpec
		},
		serviceMetadata: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.Metadata.CDAPServiceSpec
		},
		serviceMetrics: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.Metrics.CDAPServiceSpec
		},
		servicePreview: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.Preview.CDAPServiceSpec
		},
		serviceRouter: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.Router.CDAPServiceSpec
		},
		serviceUserInterface: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPServiceSpec {
			return &master.Spec.UserInterface.CDAPServiceSpec
		},
	}
	return serviceSpecMap[serviceName](master)
}

func getCDAPStatefulServiceSpec(serviceName ServiceName, master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
	serviceSpecMap := map[ServiceName]func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec{
		serviceAppFabric: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return nil
		},
		serviceLogs: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return &master.Spec.Logs.CDAPStatefulServiceSpec
		},
		serviceMessaging: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return &master.Spec.Messaging.CDAPStatefulServiceSpec
		},
		serviceMetadata: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return nil
		},
		serviceMetrics: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return &master.Spec.Metrics.CDAPStatefulServiceSpec
		},
		servicePreview: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return &master.Spec.Preview.CDAPStatefulServiceSpec
		},
		serviceRouter: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return nil
		},
		serviceUserInterface: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPStatefulServiceSpec {
			return nil
		},
	}
	return serviceSpecMap[serviceName](master)
}

func getCDAPExternalService(serviceName ServiceName, master *v1alpha1.CDAPMaster) *v1alpha1.CDAPExternalServiceSpec {
	serviceSpecMap := map[ServiceName]func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPExternalServiceSpec{
		serviceRouter: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPExternalServiceSpec {
			return &master.Spec.Router.CDAPExternalServiceSpec
		},
		serviceUserInterface: func(master *v1alpha1.CDAPMaster) *v1alpha1.CDAPExternalServiceSpec {
			return &master.Spec.UserInterface.CDAPExternalServiceSpec
		},
	}
	return serviceSpecMap[serviceName](master)
}
