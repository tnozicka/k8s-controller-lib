package resourceapply

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tnozicka/k8s-controller-lib/pkg/controllerhelpers"
)

type HashAnnotator interface {
	GetHashAnnotationKey() string
	GetHashAnnotation(obj metav1.Object) string
	SetHashAnnotationWithCleanup(obj metav1.Object) error
}

type hashAnnotator struct {
	annotationKey string
}

var _ HashAnnotator = &hashAnnotator{}

func NewHashAnnotator(annotationKey string) HashAnnotator {
	return &hashAnnotator{
		annotationKey: annotationKey,
	}
}

func (h *hashAnnotator) GetHashAnnotationKey() string {
	return h.annotationKey
}

func (h *hashAnnotator) GetHashAnnotation(obj metav1.Object) string {
	return obj.GetAnnotations()[h.annotationKey]
}

func (h *hashAnnotator) SetHashAnnotationWithCleanup(obj metav1.Object) error {
	obj.SetUID("")
	obj.SetCreationTimestamp(metav1.Time{})
	obj.SetGeneration(0)
	obj.SetManagedFields(nil)
	obj.SetSelfLink("")

	// We do not want to hash ResourceVersion, but we need to preserve it for the update semantics.
	rv := obj.GetResourceVersion()
	obj.SetResourceVersion("")
	defer obj.SetResourceVersion(rv)

	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Clear the annotation that holds the hash, so this produces a consistent output.
	delete(annotations, h.annotationKey)

	hash, err := controllerhelpers.HashObjects(obj)
	if err != nil {
		return fmt.Errorf("can't hash objects: %w", err)
	}

	annotations[h.annotationKey] = hash
	obj.SetAnnotations(annotations)

	return nil
}
