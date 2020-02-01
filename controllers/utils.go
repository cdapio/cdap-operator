package controllers

import (
	"fmt"
	"sigs.k8s.io/controller-reconciler/pkg/reconciler"
	"strings"
)

// Creates a int32 pointer for the given value
func int32Ptr(value int32) *int32 {
	return &value
}

func getObjName(masterName, name string) string {
	return fmt.Sprintf("%s%s-%s", objectNamePrefix, masterName, strings.ToLower(name))
}

func getServiceMain(name ServiceName) string {
	return fmt.Sprintf("%sServiceMain", name)
}

func getReplicas(replicas *int32) int32 {
	var r int32 = 1
	if replicas != nil {
		r = *replicas
	}
	return r
}

func mergeLabels(current, added map[string]string) map[string]string {
	labels := make(reconciler.KVMap)
	labels.Merge(current, added)
	return labels
}
