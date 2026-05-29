package kubernetes

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	corev1listers "k8s.io/client-go/listers/core/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyNamespaceWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*corev1.Namespace],
	required *corev1.Namespace,
	options resourceapply.ApplyOptions,
) (*corev1.Namespace, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyNamespace(
	ctx context.Context,
	a *resourceapply.Applier,
	client corev1client.NamespacesGetter,
	lister corev1listers.NamespaceLister,
	required *corev1.Namespace,
	options resourceapply.ApplyOptions,
) (*corev1.Namespace, bool, error) {
	return ApplyNamespaceWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*corev1.Namespace]{
			GetCached: lister.Get,
			Create:    client.Namespaces().Create,
			Update:    client.Namespaces().Update,
			Delete:    client.Namespaces().Delete,
		},
		required,
		options,
	)
}

func ApplyServiceAccountWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*corev1.ServiceAccount],
	required *corev1.ServiceAccount,
	options resourceapply.ApplyOptions,
) (*corev1.ServiceAccount, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyServiceAccount(
	ctx context.Context,
	a *resourceapply.Applier,
	client corev1client.ServiceAccountsGetter,
	lister corev1listers.ServiceAccountLister,
	required *corev1.ServiceAccount,
	options resourceapply.ApplyOptions,
) (*corev1.ServiceAccount, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyServiceAccountWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*corev1.ServiceAccount]{
			GetCached: lister.ServiceAccounts(ns).Get,
			Create:    client.ServiceAccounts(ns).Create,
			Update:    client.ServiceAccounts(ns).Update,
			Delete:    client.ServiceAccounts(ns).Delete,
		},
		required,
		options,
	)
}

func ApplyServiceWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*corev1.Service],
	required *corev1.Service,
	options resourceapply.ApplyOptions,
) (*corev1.Service, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyService(
	ctx context.Context,
	a *resourceapply.Applier,
	client corev1client.ServicesGetter,
	lister corev1listers.ServiceLister,
	required *corev1.Service,
	options resourceapply.ApplyOptions,
) (*corev1.Service, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyServiceWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*corev1.Service]{
			GetCached: lister.Services(ns).Get,
			Create:    client.Services(ns).Create,
			Update:    client.Services(ns).Update,
			Delete:    client.Services(ns).Delete,
		},
		required,
		options,
	)
}

func ApplySecretWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*corev1.Secret],
	required *corev1.Secret,
	options resourceapply.ApplyOptions,
) (*corev1.Secret, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplySecret(
	ctx context.Context,
	a *resourceapply.Applier,
	client corev1client.SecretsGetter,
	lister corev1listers.SecretLister,
	required *corev1.Secret,
	options resourceapply.ApplyOptions,
) (*corev1.Secret, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplySecretWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*corev1.Secret]{
			GetCached: lister.Secrets(ns).Get,
			Create:    client.Secrets(ns).Create,
			Update:    client.Secrets(ns).Update,
			Delete:    client.Secrets(ns).Delete,
		},
		required,
		options,
	)
}

func ApplyConfigMapWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*corev1.ConfigMap],
	required *corev1.ConfigMap,
	options resourceapply.ApplyOptions,
) (*corev1.ConfigMap, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyConfigMap(
	ctx context.Context,
	a *resourceapply.Applier,
	client corev1client.ConfigMapsGetter,
	lister corev1listers.ConfigMapLister,
	required *corev1.ConfigMap,
	options resourceapply.ApplyOptions,
) (*corev1.ConfigMap, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyConfigMapWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*corev1.ConfigMap]{
			GetCached: lister.ConfigMaps(ns).Get,
			Create:    client.ConfigMaps(ns).Create,
			Update:    client.ConfigMaps(ns).Update,
			Delete:    client.ConfigMaps(ns).Delete,
		},
		required,
		options,
	)
}

func ApplyResourceQuotaWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*corev1.ResourceQuota],
	required *corev1.ResourceQuota,
	options resourceapply.ApplyOptions,
) (*corev1.ResourceQuota, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyResourceQuota(
	ctx context.Context,
	a *resourceapply.Applier,
	client corev1client.ResourceQuotasGetter,
	lister corev1listers.ResourceQuotaLister,
	required *corev1.ResourceQuota,
	options resourceapply.ApplyOptions,
) (*corev1.ResourceQuota, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyResourceQuotaWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*corev1.ResourceQuota]{
			GetCached: lister.ResourceQuotas(ns).Get,
			Create:    client.ResourceQuotas(ns).Create,
			Update:    client.ResourceQuotas(ns).Update,
			Delete:    client.ResourceQuotas(ns).Delete,
		},
		required,
		options,
	)
}
