package kubernetes

import (
	"context"
	"errors"

	policyv1 "k8s.io/api/policy/v1"
	policyv1client "k8s.io/client-go/kubernetes/typed/policy/v1"
	policyv1listers "k8s.io/client-go/listers/policy/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyPodDisruptionBudgetWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*policyv1.PodDisruptionBudget],
	required *policyv1.PodDisruptionBudget,
	options resourceapply.ApplyOptions,
) (*policyv1.PodDisruptionBudget, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyPodDisruptionBudget(
	ctx context.Context,
	a *resourceapply.Applier,
	client policyv1client.PodDisruptionBudgetsGetter,
	lister policyv1listers.PodDisruptionBudgetLister,
	required *policyv1.PodDisruptionBudget,
	options resourceapply.ApplyOptions,
) (*policyv1.PodDisruptionBudget, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyPodDisruptionBudgetWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*policyv1.PodDisruptionBudget]{
			GetCached: lister.PodDisruptionBudgets(ns).Get,
			Create:    client.PodDisruptionBudgets(ns).Create,
			Update:    client.PodDisruptionBudgets(ns).Update,
		},
		required,
		options,
	)
}
