package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	v1alpha1 "cdap.io/cdap-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	labelInstanceKey = "cdap.instance"
	// TODO: add this an an extra label in kubetwillpreparer.
	cdapMasterNamespaceKey = "cdapmaster.namespace"
)

type PodMutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func NewPodMutator(client client.Client) *PodMutator {
	return &PodMutator{
		Client: client,
	}
}

func (s *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := s.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	log.Printf("Got admission request for pod name: %s", pod.Name)

	cdapMasterName := pod.ObjectMeta.Labels[labelInstanceKey]
	cdapMaster, err := s.cdapMaster(ctx, cdapMasterName, cdapMasterNamespace(pod))
	if err != nil {
		log.Printf("Error while getting cdap master: %v", err)
		if errors.IsNotFound(err) {
			return admission.Denied(fmt.Sprintf("Pod label %q refers to a non-existent CDAPMaster object %q", labelInstanceKey, cdapMasterName))
		}
		return admission.Errored(http.StatusInternalServerError, fmt.Errorf("Error while getting cdap master: %v", err))
	}

	mutationConfigs := cdapMaster.Spec.MutationConfigs
	for _, mc := range mutationConfigs {
		selector, err := metav1.LabelSelectorAsSelector(&mc.LabelSelector)
		if err != nil {
			log.Printf("Ignoring failure to parse label selector: %v", err)
			continue
		}
		if !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		// Apply pod mutations.
		if mc.PodMutations.NodeSelector != nil {
			if pod.Spec.NodeSelector == nil {
				pod.Spec.NodeSelector = map[string]string{}
			}
			for key, value := range *mc.PodMutations.NodeSelector {
				pod.Spec.NodeSelector[key] = value
			}
		}

		pod.Spec.Tolerations = append(pod.Spec.Tolerations, mc.PodMutations.Tolerations...)

		ics := append([]corev1.Container{}, mc.PodMutations.InitContainersBefore...)
		pod.Spec.InitContainers = append(ics, pod.Spec.InitContainers...)
	}

	marshaledPod, err := json.Marshal(pod)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, marshaledPod)
}

func (s *PodMutator) InjectDecoder(d *admission.Decoder) error {
	s.decoder = d
	return nil
}

func (s *PodMutator) cdapMaster(ctx context.Context, cdapMasterName, namespace string) (*v1alpha1.CDAPMaster, error) {
	log.Printf("Fetching CDAP CR with name %q.", cdapMasterName)
	var cdapMaster v1alpha1.CDAPMaster
	err := s.Client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: cdapMasterName}, &cdapMaster)
	if err != nil {
		return nil, err
	}
	return &cdapMaster, nil
}

func cdapMasterNamespace(pod *corev1.Pod) string {
	// First look for the cdap root namespace label.
	if ns, ok := pod.Labels[cdapMasterNamespaceKey]; ok {
		return ns
	}
	// Assume that the pod is in the same namespace as the CR.
	return pod.Namespace
}
