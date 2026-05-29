package kubernetes

import (
	"context"
	"errors"

	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1client "k8s.io/client-go/kubernetes/typed/rbac/v1"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyClusterRoleWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*rbacv1.ClusterRole],
	required *rbacv1.ClusterRole,
	options resourceapply.ApplyOptions,
) (*rbacv1.ClusterRole, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyClusterRole(
	ctx context.Context,
	a *resourceapply.Applier,
	client rbacv1client.ClusterRolesGetter,
	lister rbacv1listers.ClusterRoleLister,
	required *rbacv1.ClusterRole,
	options resourceapply.ApplyOptions,
) (*rbacv1.ClusterRole, bool, error) {
	return ApplyClusterRoleWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*rbacv1.ClusterRole]{
			GetCached: lister.Get,
			Create:    client.ClusterRoles().Create,
			Update:    client.ClusterRoles().Update,
			Delete:    client.ClusterRoles().Delete,
		},
		required,
		options,
	)
}

func ApplyClusterRoleBindingWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*rbacv1.ClusterRoleBinding],
	required *rbacv1.ClusterRoleBinding,
	options resourceapply.ApplyOptions,
) (*rbacv1.ClusterRoleBinding, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyClusterRoleBinding(
	ctx context.Context,
	a *resourceapply.Applier,
	client rbacv1client.ClusterRoleBindingsGetter,
	lister rbacv1listers.ClusterRoleBindingLister,
	required *rbacv1.ClusterRoleBinding,
	options resourceapply.ApplyOptions,
) (*rbacv1.ClusterRoleBinding, bool, error) {
	return ApplyClusterRoleBindingWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*rbacv1.ClusterRoleBinding]{
			GetCached: lister.Get,
			Create:    client.ClusterRoleBindings().Create,
			Update:    client.ClusterRoleBindings().Update,
			Delete:    client.ClusterRoleBindings().Delete,
		},
		required,
		options,
	)
}

func ApplyRoleWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*rbacv1.Role],
	required *rbacv1.Role,
	options resourceapply.ApplyOptions,
) (*rbacv1.Role, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyRole(
	ctx context.Context,
	a *resourceapply.Applier,
	client rbacv1client.RolesGetter,
	lister rbacv1listers.RoleLister,
	required *rbacv1.Role,
	options resourceapply.ApplyOptions,
) (*rbacv1.Role, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyRoleWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*rbacv1.Role]{
			GetCached: lister.Roles(ns).Get,
			Create:    client.Roles(ns).Create,
			Update:    client.Roles(ns).Update,
			Delete:    client.Roles(ns).Delete,
		},
		required,
		options,
	)
}

func ApplyRoleBindingWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*rbacv1.RoleBinding],
	required *rbacv1.RoleBinding,
	options resourceapply.ApplyOptions,
) (*rbacv1.RoleBinding, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyRoleBinding(
	ctx context.Context,
	a *resourceapply.Applier,
	client rbacv1client.RoleBindingsGetter,
	lister rbacv1listers.RoleBindingLister,
	required *rbacv1.RoleBinding,
	options resourceapply.ApplyOptions,
) (*rbacv1.RoleBinding, bool, error) {
	ns := required.Namespace
	if len(ns) == 0 {
		return nil, false, errors.New("missing namespace")
	}
	return ApplyRoleBindingWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*rbacv1.RoleBinding]{
			GetCached: lister.RoleBindings(ns).Get,
			Create:    client.RoleBindings(ns).Create,
			Update:    client.RoleBindings(ns).Update,
			Delete:    client.RoleBindings(ns).Delete,
		},
		required,
		options,
	)
}
