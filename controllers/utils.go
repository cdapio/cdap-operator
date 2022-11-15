package controllers

import (
	"fmt"
	"reflect"
	"strings"

	"cdap.io/cdap-operator/api/v1alpha1"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
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

// cloneMap creates a clone of the given source map.
func cloneMap(source map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range source {
		result[k] = v
	}
	return result
}

func mergeMaps(current, added map[string]string) map[string]string {
	labels := make(reconciler.KVMap)
	labels.Merge(current, added)
	return labels
}

// getCDAPMasterServiceSpec returns a pointer to the specification of the given service
func getCDAPMasterServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (interface{}, error) {
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
	return val.Interface(), nil
}

// getCDAPMasterSpec returns a pointer to the given specType for the given service (using reflect).
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to find the field of the given specType for the given service if allowMissingSpec is false
func getCDAPMasterSpec(master *v1alpha1.CDAPMaster, service ServiceName, specType reflect.Type, allowMissingSpec bool) (interface{}, error) {
	serviceSpec, err := getCDAPMasterServiceSpec(master, service)
	if err != nil {
		return nil, err
	}
	if serviceSpec == nil {
		return nil, nil
	}

	fieldValue, err := getFieldValue(serviceSpec, func(field reflect.StructField) bool {
		// Since we are looking for embedded type, the field name and field type must match with the specType
		return field.Type == specType && field.Name == specType.Name()
	})
	if err != nil || (fieldValue == nil && !allowMissingSpec) {
		return nil, fmt.Errorf("failed to find spec field %v for service %v", specType.Name(), service)
	}
	if fieldValue == nil {
		return nil, nil
	}

	return fieldValue.Interface(), nil
}

// Return pointer to CDAPServiceSpec for the given service (using reflect).
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to find CDAPServiceSpec field
// - unable to cast to CDAPServiceSpec type
func getCDAPServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPServiceSpec, error) {
	if spec, err := getCDAPMasterSpec(master, service, reflect.TypeOf(v1alpha1.CDAPServiceSpec{}), false); err != nil {
		return nil, err
	} else if spec == nil {
		return nil, nil
	} else {
		// The cast should succeed as it has been checked in the getCDAPMasterSpec method
		return spec.(*v1alpha1.CDAPServiceSpec), nil
	}
}

// Return pointer to CDAPScalableServiceSpec for the given service (using reflect) if it contains one, otherwise nil
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to cast to CDAPScalableServiceSpec type
func getCDAPScalableServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPScalableServiceSpec, error) {
	if spec, err := getCDAPMasterSpec(master, service, reflect.TypeOf(v1alpha1.CDAPScalableServiceSpec{}), true); err != nil {
		return nil, err
	} else if spec == nil {
		return nil, nil
	} else {
		// The cast should succeed as it has been checked in the getCDAPMasterSpec method
		return spec.(*v1alpha1.CDAPScalableServiceSpec), nil
	}
}

// Return pointer to CDAPStatefulServiceSpec for the given service (using reflect) if it contains one, otherwise nil
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to cast to CDAPStatefulServiceSpec type
func getCDAPStatefulServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPStatefulServiceSpec, error) {
	if spec, err := getCDAPMasterSpec(master, service, reflect.TypeOf(v1alpha1.CDAPStatefulServiceSpec{}), true); err != nil {
		return nil, err
	} else if spec == nil {
		return nil, nil
	} else {
		// The cast should succeed as it has been checked in the getCDAPMasterSpec method
		return spec.(*v1alpha1.CDAPStatefulServiceSpec), nil
	}
}

// Return pointer to CDAPExternalServiceSpec for the given service (using reflect) if it contains one, otherwise nil
// Fail if any of the following occurs
// - unable to find the field for the service
// - unable to cast to CDAPExternalServiceSpec type
func getCDAPExternalServiceSpec(master *v1alpha1.CDAPMaster, service ServiceName) (*v1alpha1.CDAPExternalServiceSpec, error) {
	if spec, err := getCDAPMasterSpec(master, service, reflect.TypeOf(v1alpha1.CDAPExternalServiceSpec{}), true); err != nil {
		return nil, err
	} else if spec == nil {
		return nil, nil
	} else {
		// The cast should succeed as it has been checked in the getCDAPMasterSpec method
		return spec.(*v1alpha1.CDAPExternalServiceSpec), nil
	}
}

// getFieldValue finds a field value in the given struct object that passes the predicate and return it as a *reflect.Value.
// It will recursively scan for all fields across all the struct embeddings.
// The Value return is always a pointer to the field value or nil if no such field exists.
func getFieldValue(obj interface{}, predicate func(reflect.StructField) bool) (*reflect.Value, error) {
	value := reflect.Indirect(reflect.ValueOf(obj))
	if value.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %v is not a struct", value)
	}

	var fields []reflect.Value
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if field.Kind() != reflect.Ptr {
			field = field.Addr()
		}

		if predicate(value.Type().Field(i)) {
			return &field, nil
		}

		if reflect.Indirect(field).Kind() == reflect.Struct {
			fields = append(fields, field)
		}
	}

	for _, field := range fields {
		if value, err := getFieldValue(field.Interface(), predicate); err == nil {
			return value, nil
		}
	}

	return nil, nil
}

func max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
