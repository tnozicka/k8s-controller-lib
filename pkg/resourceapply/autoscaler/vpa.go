package autoscaler

import (
	"context"
	"errors"
	"fmt"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kautoscalingv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	kautoscalingv1client "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	kautoscalingv1listers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	"k8s.io/client-go/discovery"

	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
	"github.com/tnozicka/k8s-controller-lib/pkg/resources"
)

func ApplyVerticalPodAutoscalerWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*kautoscalingv1.VerticalPodAutoscaler],
	required *kautoscalingv1.VerticalPodAutoscaler,
	options resourceapply.ApplyOptions,
) (*kautoscalingv1.VerticalPodAutoscaler, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyVerticalPodAutoscaler(
	ctx context.Context,
	a *resourceapply.Applier,
	client kautoscalingv1client.VerticalPodAutoscalersGetter,
	lister kautoscalingv1listers.VerticalPodAutoscalerLister,
	required *kautoscalingv1.VerticalPodAutoscaler,
	options resourceapply.ApplyOptions,
) (*kautoscalingv1.VerticalPodAutoscaler, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyVerticalPodAutoscalerWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*kautoscalingv1.VerticalPodAutoscaler]{
			GetCached: lister.VerticalPodAutoscalers(ns).Get,
			Create:    client.VerticalPodAutoscalers(ns).Create,
			Update:    client.VerticalPodAutoscalers(ns).Update,
		},
		required,
		options,
	)
}

func ApplyVerticalPodAutoscalerWithStatus(
	ctx context.Context,
	a *resourceapply.Applier,
	client kautoscalingv1client.VerticalPodAutoscalersGetter,
	lister kautoscalingv1listers.VerticalPodAutoscalerLister,
	required *kautoscalingv1.VerticalPodAutoscaler,
	options resourceapply.ApplyOptions,
	hasStatusSubresource bool,
) (*kautoscalingv1.VerticalPodAutoscaler, bool, error) {
	vpa, changed, err := ApplyVerticalPodAutoscaler(
		ctx,
		a,
		client,
		lister,
		required,
		options,
	)
	if err != nil {
		return nil, false, fmt.Errorf("can't update vpa %q: %w", naming.ObjNN(required), err)
	}

	if !hasStatusSubresource {
		// Status subresource is not present in older versions of the API, so the status change has already been applied.
		return vpa, changed, nil
	}

	if apiequality.Semantic.DeepEqual(required.Status, vpa.Status) {
		return vpa, changed, nil
	}

	required = required.DeepCopy()
	required.ResourceVersion = vpa.ResourceVersion
	vpa, err = client.VerticalPodAutoscalers(required.Namespace).UpdateStatus(ctx, required, metav1.UpdateOptions{
		FieldValidation: options.GetFieldValidation(),
	})
	if err != nil {
		return nil, false, fmt.Errorf("can't update status for vpa %q: %w", naming.ObjNN(required), err)
	}

	return vpa, true, nil
}

func ApplyVerticalPodAutoscalerWithStatusDiscovery(
	ctx context.Context,
	a *resourceapply.Applier,
	client kautoscalingv1client.VerticalPodAutoscalersGetter,
	lister kautoscalingv1listers.VerticalPodAutoscalerLister,
	required *kautoscalingv1.VerticalPodAutoscaler,
	options resourceapply.ApplyOptions,
	discoveryClient discovery.DiscoveryInterface,
) (*kautoscalingv1.VerticalPodAutoscaler, bool, error) {
	groupResourceList, err := discoveryClient.ServerResourcesForGroupVersion(kautoscalingv1.SchemeGroupVersion.Identifier())
	if err != nil {
		return nil, false, fmt.Errorf("can't discover resources in group %q: %w", kautoscalingv1.SchemeGroupVersion, err)
	}

	allResources, err := resources.FromAPIList(groupResourceList, nil)
	if err != nil {
		return nil, false, fmt.Errorf("can't convert resource list: %w", err)
	}

	hasStatusSubresource := allResources.Has(kautoscalingv1.SchemeGroupVersion.WithResource("verticalpodautoscalers/status"))
	return ApplyVerticalPodAutoscalerWithStatus(
		ctx,
		a,
		client,
		lister,
		required,
		options,
		hasStatusSubresource,
	)
}
