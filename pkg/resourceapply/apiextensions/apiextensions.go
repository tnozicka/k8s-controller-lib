package apiextensions

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1"
	apiextensionsv1listers "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/resourceapply"
)

func ApplyCustomResourceDefinitionWithControl(
	ctx context.Context,
	a *resourceapply.Applier,
	control resourceapply.ApplyControl[*apiextensionsv1.CustomResourceDefinition],
	required *apiextensionsv1.CustomResourceDefinition,
	options resourceapply.ApplyOptions,
) (*apiextensionsv1.CustomResourceDefinition, bool, error) {
	return resourceapply.ApplyGeneric(
		ctx,
		a,
		control,
		required,
		options,
	)
}

func ApplyCustomResourceDefinition(
	ctx context.Context,
	a *resourceapply.Applier,
	client apiextensionsv1client.CustomResourceDefinitionsGetter,
	lister apiextensionsv1listers.CustomResourceDefinitionLister,
	required *apiextensionsv1.CustomResourceDefinition,
	options resourceapply.ApplyOptions,
) (*apiextensionsv1.CustomResourceDefinition, bool, error) {
	return ApplyCustomResourceDefinitionWithControl(
		ctx,
		a,
		resourceapply.ApplyControl[*apiextensionsv1.CustomResourceDefinition]{
			GetCached: lister.Get,
			Create:    client.CustomResourceDefinitions().Create,
			Update:    client.CustomResourceDefinitions().Update,
			Delete:    client.CustomResourceDefinitions().Delete,
		},
		required,
		options,
	)
}
