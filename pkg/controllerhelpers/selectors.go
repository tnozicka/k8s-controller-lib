package controllerhelpers

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/pkg/kubetypes"
	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type DeleteFunc func(ctx context.Context, name string, opts metav1.DeleteOptions) error

type ObjectMapTracker interface {
	Len() int
	ObjectTypeString() string
	DeleteObjects(ctx context.Context) error
}

type objectMapTracker[K comparable, O kubetypes.Object] struct {
	m            *map[K]O
	deleteObject DeleteFunc
}

func NewObjectMapTracker[K comparable, O kubetypes.Object](m *map[K]O, deleteObject DeleteFunc) ObjectMapTracker {
	return objectMapTracker[K, O]{
		m:            m,
		deleteObject: deleteObject,
	}
}

var _ ObjectMapTracker = objectMapTracker[string, kubetypes.Object]{}

func (omt objectMapTracker[K, O]) Len() int {
	return len(*omt.m)
}

func (omt objectMapTracker[K, O]) ObjectTypeString() string {
	return fmt.Sprintf("%T", *new(O))
}

func (omt objectMapTracker[K, O]) DeleteObjects(ctx context.Context) error {
	for _, obj := range *omt.m {
		klog.V(2).InfoS(
			"Deleting remaining object",
			"Object", klog.KObj(obj),
			"GoType", omt.ObjectTypeString(),
		)
		err := omt.deleteObject(ctx, obj.GetName(), metav1.DeleteOptions{
			Preconditions: metav1.NewUIDPreconditions(string(obj.GetUID())),
		})
		if err != nil {
			return fmt.Errorf("can't delete object %q (%T): %w", naming.ObjNN(obj), obj, err)
		}
	}

	return nil
}

func FilterOutObjectMapBySelector[T metav1.Object](objects *map[string]T, selector labels.Selector) map[string]T {
	res := map[string]T{}

	for name, obj := range *objects {
		if selector.Matches(labels.Set(obj.GetLabels())) {
			res[name] = obj
			delete(*objects, name)
		}
	}

	return res
}
