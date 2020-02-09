package controllers

import (
	//	"reflect"

	//	alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	"fmt"

	//"cdap.io/cdap-operator/api/v1alpha1"
	"encoding/json"
	"io/ioutil"
	"testing"
)

func jsonUnmarshal(filename string, obj interface{}) error {
	file, _ := ioutil.ReadFile(filename)
	if err := json.Unmarshal([]byte(file), obj); err != nil {
		return fmt.Errorf("failed to read expected spec json file %v", err)
	}
	return nil
}

func TestCDAPDeploymentSpec(t *testing.T) {
	////master := &v1alpha1.CDAPMaster{}
	////labels := mergeMaps(master.Labels, map[string]string{"using": "controllers.ServiceHandler"})
	////_, err := buildCDAPDeploymentSpec(master, labels)
	////if err != nil {
	////	t.Errorf("build CDAP deployment spec failed %v", err)
	////}
	//
	//master := v1alpha1.CDAPMaster{}
	//if err := jsonUnmarshal("testdata/cdap_cr_3_pods", &master); err != nil {
	//	t.Errorf("jsonUnmarshal failed %v", err)
	//}
	//labels := mergeMaps(master.Labels, map[string]string{
	//	"using":                     "controllers.ServiceHandler",
	//	"custom-resource":           "v1alpha1.CDAPMaster",
	//	"custom-resource-name":      "test",
	//	"custom-resource-namespace": "default",
	//})
	//spec, err := buildCDAPDeploymentSpec(&master, labels)
	//if err != nil {
	//	t.Errorf("build CDAP deployment spec failed %v", err)
	//}
	//
	//expectedSpec := &CDAPDeploymentSpec{}
	//if err := jsonUnmarshal("testdata/cdap_deployment_spec_3_pods.json", expectedSpec); err != nil {
	//	t.Errorf("jsonUnmarshal failed %v", err)
	//}
	//t.Log(spec.toString())
	//t.Log(expectedSpec.toString())
	////if !reflect.DeepEqual(spec, expectedSpec) {
	////	t.Errorf("not the same")
	////}
	//if diff := deep.Equal(spec, expectedSpec); diff != nil {
	//	t.Error(diff)
	//}
}
