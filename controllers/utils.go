package controllers

import (
	"cdap.io/cdap-operator/api/v1alpha1"
	"fmt"
	"reflect"
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

func getObjName(master *v1alpha1.CDAPMaster, name string) string {
	return fmt.Sprintf("%s%s-%s", objectNamePrefix, master.Name, strings.ToLower(name))
}

func mergeMaps(current, added map[string]string) map[string]string {
	labels := make(reconciler.KVMap)
	labels.Merge(current, added)
	return labels
}
// Return pointer to CDAPServiceSpec for the given service (using reflection).
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to find CDAPServiceSpec field
// - unable to cast to CDAPServiceSpec type
func getCDAPServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPServiceSpec, error) {
	val := reflect.Indirect(reflect.ValueOf(&master.Spec)).FieldByName(service)
	if !val.IsValid() {
		return nil, fmt.Errorf("failed to find field %v in %v", service, reflect.TypeOf(master.Spec).Name())
	}
	val = reflect.Indirect(reflect.ValueOf(val.Addr().Interface())).FieldByName(fieldNameCDAPServiceSpec)
	if !val.IsValid() {
		return nil, fmt.Errorf("failed to find field %v in %v", fieldNameCDAPServiceSpec, reflect.TypeOf(val).Name())
	}
	ret, ok := val.Addr().Interface().(*v1alpha1.CDAPServiceSpec)
	if !ok {
		return nil, fmt.Errorf("failed to cast to poiter to CDAPServiceSpec")
	}
	return ret, nil
}

// TODO: simplify the code by using reflection
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

// TODO: simplify the code by using reflection
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

func min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}