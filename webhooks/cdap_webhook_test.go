package webhooks

import (
	"context"
	"fmt"
	"testing"

	"cdap.io/cdap-operator/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMutatePod(t *testing.T) {
	ctx := context.Background()

	cdapMaster := &v1alpha1.CDAPMaster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cdap-instance-1",
			Namespace: "cdap-namespace",
		},
		Spec: v1alpha1.CDAPMasterSpec{
			MutationConfigs: []v1alpha1.MutationConfig{
				{
					LabelSelector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "cdap.twill.app",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"app-1", "app-2", "app-3"},
							},
						},
					},
					PodMutations: v1alpha1.PodMutationConfig{
						NodeSelector: &map[string]string{
							"secure-nodepool": "true",
						},
						InitContainersBefore: []v1.Container{
							{
								Name: "init-container-1",
							},
						},
						Tolerations: []v1.Toleration{
							{
								Key:   "toleration-key",
								Value: "toleration-value",
							},
						},
					},
				},
			},
		},
	}

	testCases := []struct {
		description string
		pod         v1.Pod
		wantErr     error
		wantPod     v1.Pod
	}{
		{
			description: "successful",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "cdap-namespace",
					Labels: map[string]string{
						"cdap.twill.app": "app-2",
						"cdap.instance":  "cdap-instance-1",
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "original-init-container",
						},
					},
					NodeSelector: map[string]string{
						"container-scanner": "true",
					},
				},
			},
			wantPod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "cdap-namespace",
					Labels: map[string]string{
						"cdap.twill.app": "app-2",
						"cdap.instance":  "cdap-instance-1",
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "init-container-1",
						},
						{
							Name: "original-init-container",
						},
					},
					Tolerations: []v1.Toleration{
						{
							Key:   "toleration-key",
							Value: "toleration-value",
						},
					},
					NodeSelector: map[string]string{
						"secure-nodepool":   "true",
						"container-scanner": "true",
					},
				},
			},
		},
		{
			description: "pod_in_other_namespace",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
					Labels: map[string]string{
						"cdap.twill.app":     "app-1",
						"cdap.instance":      "cdap-instance-1",
						"cdap.k8s.namespace": "cdap-namespace",
					},
				},
			},
			wantPod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
					Labels: map[string]string{
						"cdap.twill.app":     "app-1",
						"cdap.instance":      "cdap-instance-1",
						"cdap.k8s.namespace": "cdap-namespace",
					},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name: "init-container-1",
						},
					},
					Tolerations: []v1.Toleration{
						{
							Key:   "toleration-key",
							Value: "toleration-value",
						},
					},
					NodeSelector: map[string]string{
						"secure-nodepool": "true",
					},
				},
			},
		},
		{
			description: "non_existent_cdap_master",
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
					Labels: map[string]string{
						"cdap.twill.app": "app-1",
						"cdap.instance":  "cdap-instance-1",
					},
				},
			},
			wantErr: cmpopts.AnyError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			client, err := newFakeClient()
			if err != nil {
				t.Fatalf("Failed to create fake client: %v", err)
			}

			if err := client.Create(ctx, cdapMaster.DeepCopy()); err != nil {
				t.Fatalf("Failed to create CDAP CR: %v", err)
			}

			webhook := PodMutator{
				Client: client,
			}

			pod := tc.pod.DeepCopy()
			err = webhook.mutatePod(ctx, pod)
			if !cmp.Equal(err, tc.wantErr, cmpopts.EquateErrors()) {
				t.Fatalf("mutatePod(%+v) returned unexpected error: want %v, got %v", tc.pod, tc.wantErr, err)
			}
			if err != nil {
				t.Logf("%v", errors.IsNotFound(err))
				return
			}
			if diff := cmp.Diff(tc.wantPod, *pod); diff != "" {
				t.Errorf("mutatePod(%+v) returned unexpected diff:(-want +got):\n%s", pod, diff)
			}
		})
	}
}

// newFakeClient returns a fake kubernetes client for unit test.
func newFakeClient() (client.WithWatch, error) {
	scheme, err := newScheme()
	if err != nil {
		return nil, err
	}
	return fake.NewClientBuilder().WithScheme(scheme).Build(), nil
}

func newScheme() (*runtime.Scheme, error) {
	sch := runtime.NewScheme()
	if err := scheme.AddToScheme(sch); err != nil {
		return nil, fmt.Errorf("failed to add scheme: %v", err)
	}

	if err := v1alpha1.AddToScheme(sch); err != nil {
		return nil, fmt.Errorf("failed to add cdap to scheme: %v", err)
	}

	metav1.AddToGroupVersion(sch, v1alpha1.GroupVersion)
	return sch, nil
}
