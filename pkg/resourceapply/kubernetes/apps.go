package kubernetes

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
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
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
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
		},
		required,
		options,
	)
}
