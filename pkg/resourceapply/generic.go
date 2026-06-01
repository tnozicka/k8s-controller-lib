package resourceapply

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/controllerhelpers"
	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
	"github.com/tnozicka/k8s-controller-lib/pkg/resourcemerge"
)

type ApplyOptions struct {
	ForceOwnership            bool
	AllowMissingControllerRef bool
	FieldValidation           *string
}

func (a ApplyOptions) GetFieldValidation() string {
	if a.FieldValidation == nil {
		return metav1.FieldValidationStrict
	}

	return *a.FieldValidation
}

type ApplyControl[T kubetypes.Object] struct {
	GetCached func(name string) (T, error)
	Create    func(ctx context.Context, obj T, opts metav1.CreateOptions) (T, error)
	Update    func(ctx context.Context, obj T, opts metav1.UpdateOptions) (T, error)
	Delete    func(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

type Applier struct {
	hashAnnotator HashAnnotator
	scheme        *runtime.Scheme
	recorder      record.EventRecorder
}

func NewApplier(
	managedHashAnnotationKey string,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Applier {
	return &Applier{
		hashAnnotator: NewHashAnnotator(managedHashAnnotationKey),
		scheme:        scheme,
		recorder:      recorder,
	}
}

func (a *Applier) getObjectGVK(obj runtime.Object) (schema.GroupVersionKind, error) {
	return kubetypes.GetObjectGVK(obj, a.scheme)
}

func (a *Applier) reportCreateEvent(obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error) {
	controllerhelpers.ReportCreateEvent(a.recorder, obj, gvk, operationErr)
}

func (a *Applier) reportUpdateEvent(obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error) {
	controllerhelpers.ReportUpdateEvent(a.recorder, obj, gvk, operationErr)
}

func (a *Applier) reportDeleteEvent(obj kubetypes.Object, gvk schema.GroupVersionKind, operationErr error) {
	controllerhelpers.ReportDeleteEvent(a.recorder, obj, gvk, operationErr)
}

type ApplyGenericHandlers[T kubetypes.Object] struct {
	GetRecreateReason func(required T, existing T) (string, *metav1.DeletionPropagation, error)
}

func ApplyGenericWithHandlers[T kubetypes.Object](
	ctx context.Context,
	applier *Applier,
	control ApplyControl[T],
	required T,
	options ApplyOptions,
	handlers ApplyGenericHandlers[T],
) (T, bool, error) {
	gvk, err := applier.getObjectGVK(required)
	if err != nil {
		return *new(T), false, fmt.Errorf("can't resolve gvk for object %T: %w", required, err)
	}

	requiredControllerRef := metav1.GetControllerOfNoCopy(required)
	if requiredControllerRef == nil && !options.AllowMissingControllerRef {
		return *new(T), false, fmt.Errorf("%q is missing controllerRef", naming.ObjKindNN(gvk, required))
	}

	requiredCopy := required.DeepCopyObject().(T)

	err = applier.hashAnnotator.SetHashAnnotationWithCleanup(requiredCopy)
	if err != nil {
		return *new(T), false, fmt.Errorf("can't annotate object %q: %w", naming.ObjKindNN(gvk, required), err)
	}

	createOptions := metav1.CreateOptions{
		FieldValidation: options.GetFieldValidation(),
	}

	existing, err := control.GetCached(requiredCopy.GetName())
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return *new(T), false, fmt.Errorf("apply can't get cached object for %q: %w", gvk, err)
		}

		var actual T
		actual, err = control.Create(
			ctx,
			requiredCopy,
			createOptions,
		)
		if apierrors.IsAlreadyExists(err) {
			klog.V(2).InfoS("Object already exists (stale cache)", "GVK", gvk, "Ref", klog.KObj(requiredCopy))
		} else if err == nil {
			applier.reportCreateEvent(requiredCopy, gvk, err)
		}
		return actual, err == nil, err
	}

	var requiredControllerRefUID types.UID
	if requiredControllerRef != nil {
		requiredControllerRefUID = requiredControllerRef.UID
	}

	existingControllerRef := metav1.GetControllerOfNoCopy(existing)
	var existingControllerRefUID types.UID
	if existingControllerRef != nil {
		existingControllerRefUID = existingControllerRef.UID
	}

	if existingControllerRef == nil && requiredControllerRef != nil && options.ForceOwnership {
		klog.V(2).InfoS(
			"Forcing apply to claim the object",
			"GVK", gvk,
			"Ref", klog.KObj(requiredCopy),
		)
	} else if existingControllerRefUID != requiredControllerRefUID {
		err = fmt.Errorf("%q isn't controlled by us", naming.ObjKindNN(gvk, requiredCopy))
		applier.reportUpdateEvent(requiredCopy, gvk, err)
		return *new(T), false, err
	}

	existingHash := existing.GetAnnotations()[applier.hashAnnotator.GetHashAnnotationKey()]
	requiredHash := requiredCopy.GetAnnotations()[applier.hashAnnotator.GetHashAnnotationKey()]

	// Matching hashes mean this version of the required object is already applied.
	if existingHash == requiredHash {
		return existing, false, nil
	}

	resourcemerge.MergeMetadataInPlace(requiredCopy, existing)

	var recreateReason string
	var propagationPolicy *metav1.DeletionPropagation
	if handlers.GetRecreateReason != nil {
		recreateReason, propagationPolicy, err = handlers.GetRecreateReason(requiredCopy, existing)
		if err != nil {
			return *new(T), false, fmt.Errorf("can't get recreate reason: %w", err)
		}
	}
	if len(recreateReason) > 0 {
		klog.V(2).InfoS(
			"Apply needs to recreate the object",
			"Reason", recreateReason,
			"GVK", gvk,
			"Ref", naming.ObjWithUID(existing),
		)

		if propagationPolicy == nil {
			propagationPolicy = new(metav1.DeletePropagationBackground)
		}

		err = control.Delete(ctx, existing.GetName(), metav1.DeleteOptions{
			PropagationPolicy: propagationPolicy,
		})
		applier.reportDeleteEvent(existing, gvk, err)
		if err != nil {
			return *new(T), false, err
		}

		requiredCopy.SetResourceVersion("")

		created, err := control.Create(ctx, requiredCopy, createOptions)
		applier.reportCreateEvent(requiredCopy, gvk, err)
		if err != nil {
			return *new(T), false, err
		}

		return created, true, nil
	}

	// We shall honor the required ResourceVersion, if it was set.
	// (Required objects use ResourceVersion in case their input is based on a previous version of themselves.)
	if len(requiredCopy.GetResourceVersion()) == 0 {
		requiredCopy.SetResourceVersion(existing.GetResourceVersion())
	}

	actual, err := control.Update(
		ctx,
		requiredCopy,
		metav1.UpdateOptions{
			FieldValidation: options.GetFieldValidation(),
		},
	)
	if apierrors.IsConflict(err) {
		klog.V(2).InfoS("Hit conflict while updating the object, will retry later.", "GVK", gvk, "Ref", klog.KObj(requiredCopy))
	} else {
		applier.reportUpdateEvent(requiredCopy, gvk, err)
	}
	if err != nil {
		return *new(T), false, fmt.Errorf("can't update object %q: %w", naming.ObjKindNN(gvk, requiredCopy), err)
	}

	return actual, true, nil
}

func ApplyGeneric[T kubetypes.Object](
	ctx context.Context,
	applier *Applier,
	control ApplyControl[T],
	required T,
	options ApplyOptions,
) (T, bool, error) {
	return ApplyGenericWithHandlers(ctx, applier, control, required, options, ApplyGenericHandlers[T]{})
}
