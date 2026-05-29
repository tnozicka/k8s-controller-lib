package kubernetes

import (
	"context"
	"errors"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	appsv1listers "k8s.io/client-go/listers/apps/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyDeploymentWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*appsv1.Deployment],
	required *appsv1.Deployment,
	options resourceapply.ApplyOptions,
) (*appsv1.Deployment, bool, error) {
	return resourceapply.ApplyGenericWithHandlers(
		ctx,
		a,
		control,
		required,
		options,
		resourceapply.ApplyGenericHandlers[*appsv1.Deployment]{
			GetRecreateReason: func(required *appsv1.Deployment, existing *appsv1.Deployment) (string, *metav1.DeletionPropagation, error) {
				if !equality.Semantic.DeepEqual(existing.Spec.Selector, required.Spec.Selector) {
					existingPodLabels := existing.Spec.Template.Labels
					requiredSelector, err := metav1.LabelSelectorAsSelector(required.Spec.Selector)
					if err != nil {
						return "", nil, fmt.Errorf("can't parse required Deployment selector: %w", err)
					}

					if !requiredSelector.Matches(labels.Set(existingPodLabels)) {
						return "", nil, fmt.Errorf("required Deployment selector %q doesn't match existing Pod Labels set %v", requiredSelector, existingPodLabels)
					}

					return "spec.selector is immutable", new(metav1.DeletePropagationOrphan), nil
				}
				return "", nil, nil
			},
		},
	)
}

func ApplyDeployment(
	ctx context.Context,
	a *resourceapply.Applier,
	client appsv1client.DeploymentsGetter,
	lister appsv1listers.DeploymentLister,
	required *appsv1.Deployment,
	options resourceapply.ApplyOptions,
) (*appsv1.Deployment, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyDeploymentWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*appsv1.Deployment]{
			GetCached: lister.Deployments(ns).Get,
			Create:    client.Deployments(ns).Create,
			Update:    client.Deployments(ns).Update,
			Delete:    client.Deployments(ns).Delete,
		},
		required,
		options,
	)
}
