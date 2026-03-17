package controllerhelpers

import (
	"context"
	"fmt"
	"maps"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/tnozicka/k8s-controller-lib/pkg/controllertools"
	"github.com/tnozicka/k8s-controller-lib/pkg/helpers/slices"
	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type GetObjectsInterface[CT, T kubetypes.Object] interface {
	controllertools.CRMControlInterface[CT, T]
	ListObjects(selector labels.Selector) ([]T, error)
}

type GetObjectsFuncs[CT, T kubetypes.Object] struct {
	controllertools.CRMControl[CT, T]
	ListObjectsFunc func(selector labels.Selector) ([]T, error)
}

func (f GetObjectsFuncs[CT, T]) ListObjects(selector labels.Selector) ([]T, error) {
	return f.ListObjectsFunc(selector)
}

var _ GetObjectsInterface[kubetypes.Object, kubetypes.Object] = GetObjectsFuncs[kubetypes.Object, kubetypes.Object]{}

// GetObjectsWithFilter is a variant of GetObjects that allows filtering objects by a predicate.
// This is primarily meant for cases where the controller is partitioned,
// like when running in a DS where each instance handles its own Node.
// Most controllers don't filter the initial object set.
func GetObjectsWithFilter[CT, T kubetypes.Object](
	ctx context.Context,
	controller metav1.Object,
	controllerGVK schema.GroupVersionKind,
	selector labels.Selector,
	filterFunc func(T) bool,
	control GetObjectsInterface[CT, T],
	scheme *runtime.Scheme,
) (map[string]T, error) {
	// List all objects. If they don't match the selector, they will be orphaned in ClaimObjects().
	allObjects, err := control.ListObjects(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("can't list all objects: %w", err)
	}

	var objects []T
	for i := range allObjects {
		if filterFunc(allObjects[i]) {
			objects = append(objects, allObjects[i])
		}
	}

	return controllertools.NewControllerRefManager[CT, T](
		controller,
		controllerGVK,
		selector,
		controllertools.CRMControl[CT, T]{
			GetControllerUncachedFunc: func(ctx context.Context, name string, opts metav1.GetOptions) (CT, error) {
				return control.GetControllerUncached(ctx, name, opts)
			},
			PatchObjectFunc: func(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (T, error) {
				return control.PatchObject(ctx, name, pt, data, opts, subresources...)
			},
		},
		scheme,
	).ClaimObjects(ctx, objects)
}

// GetObjects claims all objects matching the given selector, in accordance with The Three Laws of Controllers.
// Objects that no longer match the selector but still have the controllerRef are released.
// See https://github.com/kubernetes/design-proposals-archive/blob/acc25e/api-machinery/controller-ref.md#the-three-laws-of-controllers
// for more details on the topic.
func GetObjects[CT, T kubetypes.Object](
	ctx context.Context,
	controller metav1.Object,
	controllerGVK schema.GroupVersionKind,
	selector labels.Selector,
	control GetObjectsInterface[CT, T],
	scheme *runtime.Scheme,
) (map[string]T, error) {
	return GetObjectsWithFilter(
		ctx,
		controller,
		controllerGVK,
		selector,
		func(T) bool {
			return true
		},
		control,
		scheme,
	)
}

func MintManagedData(
	obj kubetypes.Object,
	namespace string,
	requiredLabels labels.Set,
	owner kubetypes.Object,
	ownerGVK schema.GroupVersionKind,
) error {
	objNamespace := obj.GetNamespace()
	if len(objNamespace) > 0 {
		if objNamespace != namespace {
			return fmt.Errorf("object namespace %q differs from minted namespace %q", objNamespace, namespace)
		}
	}
	obj.SetNamespace(namespace)

	labelMap := obj.GetLabels()
	if labelMap == nil {
		labelMap = map[string]string{}
	}
	maps.Copy(labelMap, requiredLabels)
	obj.SetLabels(labelMap)

	existingCR := metav1.GetControllerOfNoCopy(obj)
	desiredCR := metav1.NewControllerRef(owner, ownerGVK)
	if !apiequality.Semantic.DeepEqual(existingCR, desiredCR) {
		if existingCR != nil {
			return fmt.Errorf("object %q already has a different controllerRef", naming.ObjNN(obj))
		}

		existingOwnerRefs := obj.GetOwnerReferences()
		existingOwnerRefs = append(existingOwnerRefs, *desiredCR)
		obj.SetOwnerReferences(existingOwnerRefs)
	}

	return nil
}

func MintManagedDataToSlice[T kubetypes.Object](
	objs []T,
	namespace string,
	requiredLabels labels.Set,
	owner kubetypes.Object,
	ownerGVK schema.GroupVersionKind,
) error {
	for _, obj := range objs {
		err := MintManagedData(obj, namespace, requiredLabels, owner, ownerGVK)
		if err != nil {
			return fmt.Errorf("can't mint managed data for object %q: %w", naming.ObjNN(obj), err)
		}
	}

	return nil
}

func ConvertNamesIntoLocalObjectReferences(names ...string) []corev1.LocalObjectReference {
	return slices.Convert(func(name string) corev1.LocalObjectReference {
		return corev1.LocalObjectReference{Name: name}
	}, names...)
}
