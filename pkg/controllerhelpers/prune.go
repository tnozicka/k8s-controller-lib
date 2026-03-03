package controllerhelpers

import (
	"context"
	"errors"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

// PruneObjects deletes objects that are not required anymore.
// existing is a map that holds mapping from obj.Name->obj. These objects shall already be claimed in accordance
// with the Three Laws of Controllers, likely using the ControllerRefManager.ClaimObjects function.
func PruneObjects[T kubetypes.Object](
	ctx context.Context,
	control kubetypes.GetterDeleter[T],
	eventRecorder record.EventRecorder,
	required []T,
	existing map[string]T,
	controllerGeneration int64,
	conditions *[]metav1.Condition,
	scheme *runtime.Scheme,
) error {
	var err error
	var deletionErrs []error

	pruningCond := metav1.Condition{
		Type:               fmt.Sprint("Pruning", naming.StatusConditionProgressing),
		Status:             metav1.ConditionFalse,
		Reason:             "AsExpected",
		Message:            "",
		ObservedGeneration: controllerGeneration,
	}

	for _, e := range existing {
		isRequired := false
		for _, r := range required {
			if e.GetName() == r.GetName() {
				isRequired = true
			}
		}
		if isRequired {
			continue
		}

		var gvk schema.GroupVersionKind
		gvk, err = kubetypes.GetObjectGVK(e, scheme)
		if err != nil {
			return fmt.Errorf("can't get GVK for object %T: %w", e, err)
		}

		pruningCond.Status = metav1.ConditionTrue
		pruningCond.Reason = "PruningObject"
		pruningCond.Message = fmt.Sprintf("Pruning object %s", naming.ObjKindNN(gvk, e))

		if e.GetDeletionTimestamp() != nil {
			continue
		}

		klog.V(2).InfoS("Pruning resource", "GVK", gvk, "Ref", klog.KObj(e))
		err = control.Delete(ctx, e.GetName(), metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{
				UID: new(e.GetUID()),
			},
			PropagationPolicy: new(metav1.DeletePropagationBackground),
		})
		ReportDeleteEvent(eventRecorder, e, gvk, err)
		if err != nil {
			if apierrors.IsNotFound(err) {
				klog.V(4).InfoS("Object is already deleted", "GVK", gvk, "Ref", klog.KObj(e))
			} else {
				deletionErrs = append(deletionErrs, fmt.Errorf("can't delete object %s: %w", naming.ObjKindNN(gvk, e), err))
			}
			continue
		}
	}
	err = errors.Join(deletionErrs...)
	if err != nil {
		return fmt.Errorf("can't prune objects: %w", err)
	}

	changed := meta.SetStatusCondition(conditions, pruningCond)
	if changed {
		klog.V(4).InfoS("Updated status conditions", "Condition", pruningCond)
	}

	return nil
}
