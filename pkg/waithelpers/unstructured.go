package waithelpers

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func WaitForUnstructuredObjectState(
	ctx context.Context,
	client dynamic.ResourceInterface,
	name string,
	options WaitForStateOptions,
	conditions ...func(*unstructured.Unstructured) (bool, error),
) (*unstructured.Unstructured, error) {
	return WaitForObjectState[*unstructured.Unstructured, *unstructured.UnstructuredList](
		ctx,
		client,
		name,
		options,
		conditions...,
	)
}

func WaitForUnstructuredObjectDeletion(
	ctx context.Context,
	client dynamic.ResourceInterface,
	nn types.NamespacedName,
	uid *types.UID,
) error {
	return WaitForObjectDeletion[*unstructured.Unstructured, *unstructured.UnstructuredList](
		ctx,
		client,
		nn,
		uid,
	)
}
