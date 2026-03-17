package controllertools

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	kubecontroller "github.com/tnozicka/k8s-controller-lib/thirdparty/k8s.io/kubernetes/pkg/controller"
)

type CRMControlInterface[CT, T kubetypes.Object] interface {
	PatchObject(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (T, error)
	GetControllerUncached(ctx context.Context, name string, opts metav1.GetOptions) (CT, error)
}

type CRMControl[CT, T kubetypes.Object] struct {
	PatchObjectFunc           func(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (T, error)
	GetControllerUncachedFunc func(ctx context.Context, name string, opts metav1.GetOptions) (CT, error)
}

func (cf CRMControl[CT, T]) PatchObject(ctx context.Context, name string, pt types.PatchType, data []byte, options metav1.PatchOptions, subresources ...string) (T, error) {
	return cf.PatchObjectFunc(ctx, name, pt, data, options, subresources...)
}

func (cf CRMControl[CT, T]) GetControllerUncached(ctx context.Context, name string, opts metav1.GetOptions) (CT, error) {
	return cf.GetControllerUncachedFunc(ctx, name, opts)
}

var _ CRMControlInterface[kubetypes.Object, kubetypes.Object] = CRMControl[kubetypes.Object, kubetypes.Object]{}

type ControllerRefManager[CT, T kubetypes.Object] struct {
	kubecontroller.BaseControllerRefManager
	controllerGVK schema.GroupVersionKind
	Control       CRMControlInterface[CT, T]
	scheme        *runtime.Scheme
}

func NewControllerRefManager[CT, T kubetypes.Object](
	controller metav1.Object,
	controllerGVK schema.GroupVersionKind,
	selector labels.Selector,
	control CRMControlInterface[CT, T],
	scheme *runtime.Scheme,
) *ControllerRefManager[CT, T] {
	crm := &ControllerRefManager[CT, T]{
		BaseControllerRefManager: kubecontroller.BaseControllerRefManager{
			Controller: controller,
			Selector:   selector,
		},
		controllerGVK: controllerGVK,
		Control:       control,
		scheme:        scheme,
	}
	crm.CanAdoptFunc = crm.canAdopt

	return crm
}

func (m *ControllerRefManager[CT, T]) canAdopt(ctx context.Context) error {
	// If any adoptions are attempted, we should first recheck for deletion with
	// an uncached read.
	var fresh CT
	var err error
	fresh, err = m.Control.GetControllerUncached(ctx, m.Controller.GetName(), metav1.GetOptions{})
	if err != nil {
		return err
	}

	if fresh.GetUID() != m.Controller.GetUID() {
		return fmt.Errorf("original %s controller %s/%s is gone: got uid %s, wanted %s", m.controllerGVK, m.Controller.GetNamespace(), m.Controller.GetName(), fresh.GetUID(), m.Controller.GetUID())
	}

	if fresh.GetDeletionTimestamp() != nil {
		return fmt.Errorf("%s %s/%s has just been deleted at %v", m.controllerGVK, m.Controller.GetNamespace(), m.Controller.GetName(), m.Controller.GetDeletionTimestamp())
	}

	return nil
}

func (m *ControllerRefManager[CT, T]) AdoptObject(ctx context.Context, obj T) error {
	err := m.CanAdopt(ctx)
	if err != nil {
		gvk := kubetypes.GetObjectGVKORUnknown(obj, m.scheme)
		return fmt.Errorf("can't adopt %s %s/%s (%s): %w", gvk, obj.GetNamespace(), obj.GetName(), obj.GetUID(), err)
	}

	// Note that ValidateOwnerReferences() will reject this patch if another
	// OwnerReference exists with controller=true.
	patchBytes, err := kubecontroller.OwnerRefControllerPatch(m.Controller, m.controllerGVK, obj.GetUID())
	if err != nil {
		return err
	}

	_, err = m.Control.PatchObject(ctx, obj.GetName(), types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("can't patch %s %s/%s (%s): %w", m.controllerGVK, obj.GetNamespace(), obj.GetName(), obj.GetUID(), err)
	}

	return nil
}

func (m *ControllerRefManager[CT, T]) ReleaseObject(ctx context.Context, obj T) error {
	gvk := kubetypes.GetObjectGVKORUnknown(obj, m.scheme)
	klog.V(2).InfoS("Removing controllerRef",
		"GVK", gvk,
		"Object", klog.KObj(obj),
		"ControllerRef", fmt.Sprintf("%s/%s:%s", m.controllerGVK.GroupVersion(), m.controllerGVK.Kind, m.Controller.GetName()),
	)

	patchBytes, err := kubecontroller.GenerateDeleteOwnerRefStrategicMergeBytes(obj.GetUID(), []types.UID{m.Controller.GetUID()})
	if err != nil {
		return err
	}

	_, err = m.Control.PatchObject(ctx, obj.GetName(), types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// If the object no longer exists, ignore it.
			klog.V(4).InfoS("Can't patch object because it's missing", "GVK", gvk, "Ref", klog.KObj(obj))
			return nil
		}

		if errors.IsInvalid(err) {
			// Invalid error will be returned in two cases:
			// 1. the Service has no owner reference
			// 2. the UID of the Service doesn't match because it was recreated
			// In both cases, the error can be ignored.
			klog.V(4).InfoS("Can't patch object because it's invalid", "GVK", gvk, "Ref", klog.KObj(obj))
			return nil
		}

		return err
	}

	return nil
}

func (m *ControllerRefManager[CT, T]) ClaimObjects(ctx context.Context, objects []T) (map[string]T, error) {
	match := func(obj metav1.Object) bool {
		return m.Selector.Matches(labels.Set(obj.GetLabels()))
	}
	adopt := func(ctx context.Context, obj metav1.Object) error {
		return m.AdoptObject(ctx, obj.(T))
	}
	release := func(ctx context.Context, obj metav1.Object) error {
		return m.ReleaseObject(ctx, obj.(T))
	}

	claimedMap := make(map[string]T, len(objects))
	var errs []error
	for _, obj := range objects {
		ok, err := m.ClaimObject(ctx, obj, match, adopt, release)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if ok {
			claimedMap[obj.GetName()] = obj
		}
	}
	return claimedMap, utilerrors.NewAggregate(errs)
}
