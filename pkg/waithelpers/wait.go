package waithelpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
)

func IsPresent[Object runtime.Object](_ Object) (bool, error) {
	return true, nil
}

type listerWatcher[ListObject runtime.Object] interface {
	List(context.Context, metav1.ListOptions) (ListObject, error)
	Watch(context.Context, metav1.ListOptions) (watch.Interface, error)
}

type traditionalListWatcher struct {
	*cache.ListWatch
}

func (traditionalListWatcher) IsWatchListSemanticsUnSupported() bool {
	return true
}

func makeListerWatcher[Object runtime.Object](
	client listerWatcher[Object],
	objName string,
) cache.ListerWatcher {
	fieldSelector := fields.Everything().String()
	if len(objName) > 0 {
		fieldSelector = fields.OneTermEqualSelector("metadata.name", objName).String()
	}
	return &traditionalListWatcher{
		ListWatch: &cache.ListWatch{
			ListWithContextFunc: UncachedListFunc(func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = fieldSelector
				return client.List(ctx, options)
			}),
			WatchFuncWithContext: func(ctx context.Context, options metav1.ListOptions) (i watch.Interface, e error) {
				options.FieldSelector = fieldSelector
				return client.Watch(ctx, options)
			},
		},
	}
}

type WaitForStateOptions struct {
	TolerateDelete bool
	Timeout        time.Duration
}

// UncachedListFunc wraps a List function to make sure the initial List call avoids watch cache on the apiserver.
// This is important to reason about the order of events.
func UncachedListFunc(
	f func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error),
) func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
	return func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
		// Transform every RV="0" into RV="" to avoid watch cache.
		// By default, informers use RV="0" that goes through the watch cache and can list older versions of the object.
		if options.ResourceVersion == "0" {
			options.ResourceVersion = ""
		}

		return f(ctx, options)
	}
}

func WaitForObjectState[Object, ListObject runtime.Object](
	ctx context.Context,
	client listerWatcher[ListObject],
	name string,
	options WaitForStateOptions,
	conditions ...func(obj Object) (bool, error),
) (Object, error) {
	if len(conditions) == 0 {
		return *new(Object), errors.New("no condition provided")
	}

	if options.Timeout > 0 {
		var ctxCancel context.CancelFunc
		ctx, ctxCancel = context.WithTimeout(ctx, options.Timeout)
		defer ctxCancel()
	}

	lw := makeListerWatcher(client, name)

	aggregatedCond := func(obj Object) (bool, error) {
		allDone := true
		for _, c := range conditions {
			var err error
			var done bool

			done, err = c(obj)
			if err != nil {
				return done, err
			}
			if !done {
				allDone = false
			}
		}
		return allDone, nil
	}

	event, err := watchtools.UntilWithSync(ctx, lw, nil, nil, func(event watch.Event) (bool, error) {
		switch t := event.Type; t {
		case watch.Added, watch.Modified:
			return aggregatedCond(event.Object.(Object))

		case watch.Error:
			return true, apierrors.FromObject(event.Object)

		case watch.Deleted:
			if options.TolerateDelete {
				return aggregatedCond(event.Object.(Object))
			}
			fallthrough

		default:
			return true, fmt.Errorf("unexpected event type %q", t)
		}
	})
	if err != nil {
		return *new(Object), err
	}

	return event.Object.(Object), nil
}

func WaitForObjectDeletion[Object, ListObject runtime.Object](
	ctx context.Context,
	client listerWatcher[ListObject],
	nn types.NamespacedName,
	uid *types.UID,
) error {
	lw := makeListerWatcher(client, nn.Name)
	_, err := watchtools.UntilWithSync(
		ctx,
		lw,
		nil,
		func(store cache.Store) (bool, error) {
			obj, exists, err := store.Get(&metav1.ObjectMeta{Namespace: nn.Namespace, Name: nn.Name})
			if err != nil {
				return true, err
			}
			if !exists {
				return true, nil
			}

			objMeta, err := meta.Accessor(obj)
			if err != nil {
				return true, errors.New("can't get object metadata")
			}

			objUID := objMeta.GetUID()
			if uid != nil && *uid != objUID {
				return true, nil
			}

			uid = &objUID

			return false, nil
		},
		func(e watch.Event) (bool, error) {
			switch t := e.Type; t {
			case watch.Added, watch.Bookmark:
				return false, nil
			case watch.Modified:
				// DeltaFIFO can return a modified event on re-list if the object is recreated in the meantime
				if e.Object.(metav1.Object).GetUID() != *uid {
					return true, nil
				}
				return false, nil
			case watch.Deleted:
				return true, nil
			case watch.Error:
				return true, apierrors.FromObject(e.Object)
			default:
				return true, fmt.Errorf("unexpected event type %v", t)
			}
		},
	)
	if err != nil {
		return err
	}

	return nil
}
