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

// Return pointer to CDAPServiceSpec for the given service (using reflect).
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to find CDAPServiceSpec field
// - unable to cast to CDAPServiceSpec type
func getCDAPServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPServiceSpec, error) {
	val := reflect.Indirect(reflect.ValueOf(&master.Spec)).FieldByName(service)
	if !val.IsValid() {
		return nil, fmt.Errorf("failed to find field %v in %v", service, reflect.TypeOf(master.Spec).Name())
	}
	// For optional service, its service field is a pointer to spec.
	if val.Kind() == reflect.Ptr {
		// Return nil if optional service is disabled (e.g. service field is nil)
		if val.IsNil() {
			return nil, nil
		}
	} else {
		val = val.Addr()
	}
	val = reflect.Indirect(reflect.ValueOf(val.Interface())).FieldByName(fieldNameCDAPServiceSpec)
	if !val.IsValid() {
		return nil, fmt.Errorf("failed to find field %v in %v", fieldNameCDAPServiceSpec, reflect.TypeOf(val).Name())
	}
	ret, ok := val.Addr().Interface().(*v1alpha1.CDAPServiceSpec)
	if !ok {
		return nil, fmt.Errorf("failed to cast to poiter to %v for %v", fieldNameCDAPServiceSpec, service)
	}
	return ret, nil
}

// Return pointer to CDAPStatefulServiceSpec for the given service (using reflect) if it contains one, otherwise nil
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to cast to CDAPStatefulServiceSpec type
func getCDAPStatefulServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPStatefulServiceSpec, error) {
	val := reflect.Indirect(reflect.ValueOf(&master.Spec)).FieldByName(service)
	if !val.IsValid() {
		return nil, fmt.Errorf("failed to find field %v in %v", service, reflect.TypeOf(master.Spec).Name())
	}
	// For optional service, its service field is a pointer to spec
	if val.Kind() == reflect.Ptr {
		// Return nil if optional service is disabled (e.g. service field is nil)
		if val.IsNil() {
			return nil, nil
		}
	} else {
		val = val.Addr()
	}
	val = reflect.Indirect(reflect.ValueOf(val.Interface())).FieldByName(fieldNameCDAPStatefulServiceSpec)
	if !val.IsValid() {
		return nil, nil
	}
	ret, ok := val.Addr().Interface().(*v1alpha1.CDAPStatefulServiceSpec)
	if !ok {
		return nil, fmt.Errorf("failed to cast to poiter to %v for %v", fieldNameCDAPStatefulServiceSpec, service)
	}
	return ret, nil
}

// Return pointer to CDAPExternalServiceSpec for the given service (using reflect) if it contains one, otherwise nil
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to cast to CDAPExternalServiceSpec type
func getCDAPExternalServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPExternalServiceSpec, error) {
	val := reflect.Indirect(reflect.ValueOf(&master.Spec)).FieldByName(service)
	if !val.IsValid() {
		return nil, fmt.Errorf("failed to find field %v in %v", service, reflect.TypeOf(master.Spec).Name())
	}
	val = reflect.Indirect(reflect.ValueOf(val.Addr().Interface())).FieldByName(fieldNameCDAPExternalServiceSpec)
	if !val.IsValid() {
		return nil, nil
	}
	ret, ok := val.Addr().Interface().(*v1alpha1.CDAPExternalServiceSpec)
	if !ok {
		return nil, fmt.Errorf("failed to cast to poiter to %v for %v", fieldNameCDAPExternalServiceSpec, service)
	}
	return ret, nil
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
