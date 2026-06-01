package kubeaggregator

import (
	"context"

	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	apiregistrationv1client "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset/typed/apiregistration/v1"
	apiregistrationv1listers "k8s.io/kube-aggregator/pkg/client/listers/apiregistration/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyAPIServiceWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*apiregistrationv1.APIService],
	required *apiregistrationv1.APIService,
	options resourceapply.ApplyOptions,
) (*apiregistrationv1.APIService, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyAPIService(
	ctx context.Context,
	a *resourceapply.Applier,
	client apiregistrationv1client.APIServicesGetter,
	lister apiregistrationv1listers.APIServiceLister,
	required *apiregistrationv1.APIService,
	options resourceapply.ApplyOptions,
) (*apiregistrationv1.APIService, bool, error) {
	return ApplyAPIServiceWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*apiregistrationv1.APIService]{
			GetCached: lister.Get,
			Create:    client.APIServices().Create,
			Update:    client.APIServices().Update,
			Delete:    client.APIServices().Delete,
		},
		required,
		options,
	)
}
