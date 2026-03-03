package kubernetes

import (
	"context"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	admissionregistrationv1client "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	admissionregistrationv1listers "k8s.io/client-go/listers/admissionregistration/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyMutatingWebhookConfigurationWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*admissionregistrationv1.MutatingWebhookConfiguration],
	required *admissionregistrationv1.MutatingWebhookConfiguration,
	options resourceapply.ApplyOptions,
) (*admissionregistrationv1.MutatingWebhookConfiguration, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyMutatingWebhookConfiguration(
	ctx context.Context,
	a *resourceapply.Applier,
	client admissionregistrationv1client.MutatingWebhookConfigurationsGetter,
	lister admissionregistrationv1listers.MutatingWebhookConfigurationLister,
	required *admissionregistrationv1.MutatingWebhookConfiguration,
	options resourceapply.ApplyOptions,
) (*admissionregistrationv1.MutatingWebhookConfiguration, bool, error) {
	return ApplyMutatingWebhookConfigurationWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*admissionregistrationv1.MutatingWebhookConfiguration]{
			GetCached: lister.Get,
			Create:    client.MutatingWebhookConfigurations().Create,
			Update:    client.MutatingWebhookConfigurations().Update,
		},
		required,
		options,
	)
}
