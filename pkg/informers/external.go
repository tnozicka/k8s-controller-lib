package informers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/tnozicka/k8s-controller-lib/pkg/naming"
)

type controllerAdapter struct {
	run                     func(ctx context.Context)
	hasSynced               func() bool
	lastSyncResourceVersion func() string
}

func (c *controllerAdapter) Run(stopCh <-chan struct{}) {
	c.RunWithContext(wait.ContextForChannel(stopCh))
}

func (c *controllerAdapter) RunWithContext(ctx context.Context) {
	c.run(ctx)
}

func (c *controllerAdapter) HasSynced() bool {
	return c.hasSynced()
}

func (c *controllerAdapter) LastSyncResourceVersion() string {
	return c.lastSyncResourceVersion()
}

type processorListener struct {
	cache.ResourceEventHandler
}

func (l processorListener) HasSynced() bool {
	return true
}

type External interface {
	cache.SharedIndexInformer
	// RunWithContext(ctx context.Context)
	Lister() cache.GenericLister
	Informer() cache.SharedIndexInformer
	Name() string
	ReplaceObjects(objects []any) error
}

type external struct {
	gvr schema.GroupVersionResource

	indexer cache.Indexer
	queue   *cache.DeltaFIFO

	listeners     sets.Set[*processorListener]
	listenersLock sync.RWMutex

	blockDeltas sync.Mutex

	started     bool
	startedLock sync.Mutex
	stopped     atomic.Bool

	transform cache.TransformFunc
}

var _ External = &external{}
var _ cache.SharedIndexInformer = &external{}
var _ informers.GenericInformer = &external{}

func NewExternal(gvr schema.GroupVersionResource, keyFunc cache.KeyFunc, transformer cache.TransformFunc) External {
	indexer := cache.NewIndexer(keyFunc, cache.Indexers{})
	return &external{
		gvr:     gvr,
		indexer: indexer,
		queue: cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
			KnownObjects:          indexer,
			EmitDeltaTypeReplaced: true,
			Transformer:           transformer,
		}),
		listeners:     sets.New[*processorListener](),
		listenersLock: sync.RWMutex{},
		blockDeltas:   sync.Mutex{},
	}
}

func (e *external) Name() string {
	return naming.ConciseGVR(e.gvr).String()
}

func (e *external) addListener(listener *processorListener) cache.ResourceEventHandlerRegistration {
	e.listenersLock.Lock()
	defer e.listenersLock.Unlock()

	e.listeners.Insert(listener)

	return listener
}

func (e *external) removeListener(handle cache.ResourceEventHandlerRegistration) error {
	e.listenersLock.Lock()
	defer e.listenersLock.Unlock()

	listener, ok := handle.(*processorListener)
	if !ok {
		return fmt.Errorf("invalid handle type %T", handle)
	}

	if !e.listeners.Has(listener) {
		return fmt.Errorf("listener %v not found", listener)
	}

	e.listeners.Delete(listener)

	return nil
}

func (e *external) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	e.startedLock.Lock()
	defer e.startedLock.Unlock()

	if e.stopped.Load() {
		return nil, fmt.Errorf("can't add handler %v to the external informer because it has stopped already", handler)
	}

	listener := &processorListener{
		ResourceEventHandler: handler,
	}

	if !e.started {
		return e.addListener(listener), nil
	}

	e.blockDeltas.Lock()
	defer e.blockDeltas.Unlock()

	reg := e.addListener(listener)

	for _, item := range e.indexer.List() {
		listener.OnAdd(item, true)
	}

	return reg, nil
}

func (e *external) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	return e.AddEventHandlerWithOptions(handler, cache.HandlerOptions{ResyncPeriod: &resyncPeriod})
}

func (e *external) AddEventHandlerWithOptions(_ cache.ResourceEventHandler, _ cache.HandlerOptions) (cache.ResourceEventHandlerRegistration, error) {
	panic("individual handler options are not available for external informers")
}

func (e *external) RemoveEventHandler(handle cache.ResourceEventHandlerRegistration) error {
	e.startedLock.Lock()
	defer e.startedLock.Unlock()

	e.blockDeltas.Lock()
	defer e.blockDeltas.Unlock()

	return e.removeListener(handle)
}

func (e *external) GetStore() cache.Store {
	return e.indexer
}

func (e *external) GetIndexer() cache.Indexer {
	return e.indexer
}

func (e *external) AddIndexers(indexers cache.Indexers) error {
	e.startedLock.Lock()
	defer e.startedLock.Unlock()

	if e.stopped.Load() {
		return errors.New("can't add indexer to the external informer because it has stopped already")
	}

	return e.indexer.AddIndexers(indexers)
}

func (e *external) GetController() cache.Controller {
	return &controllerAdapter{
		run: func(ctx context.Context) {
			<-ctx.Done()
		},
		hasSynced:               e.HasSynced,
		lastSyncResourceVersion: e.LastSyncResourceVersion,
	}
}

func (e *external) ReplaceObjects(objects []any) error {
	e.blockDeltas.Lock()
	defer e.blockDeltas.Unlock()

	err := e.queue.Replace(objects, "")
	if err != nil {
		return fmt.Errorf("can't replace objects: %w", err)
	}

	return nil
}

func (e *external) sync(_ context.Context, deltas cache.Deltas, isInInitialList bool) error {
	e.listenersLock.RLock()
	defer e.listenersLock.RUnlock()

	var err error
	for _, d := range deltas {
		switch d.Type {
		case cache.Sync, cache.Replaced, cache.Added, cache.Updated:
			var old any
			var exists bool
			old, exists, err = e.indexer.Get(d.Object)
			if err != nil {
				return fmt.Errorf("can't get object: %w", err)
			}

			if exists {
				err = e.indexer.Update(d.Object)
				if err != nil {
					return fmt.Errorf("can't update store: %w", err)
				}

				for l := range e.listeners {
					l.OnUpdate(old, d.Object)
				}
			} else {
				err = e.indexer.Add(d.Object)
				if err != nil {
					return fmt.Errorf("can't add object to store: %w", err)
				}

				for l := range e.listeners {
					l.OnAdd(d.Object, isInInitialList)
				}
			}

		case cache.Deleted:
			err = e.indexer.Delete(d.Object)
			if err != nil {
				return fmt.Errorf("can't delete object from store: %w", err)
			}

			for l := range e.listeners {
				l.OnDelete(d.Object)
			}
		}
	}

	return nil
}

func (e *external) processLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := e.queue.Pop(cache.PopProcessFunc(func(obj any, isInInitialList bool) error {
				deltas, ok := obj.(cache.Deltas)
				if !ok {
					return fmt.Errorf("invalid object type %T", obj)
				}

				return e.sync(ctx, deltas, isInInitialList)
			}))
			if err != nil {
				if errors.Is(err, cache.ErrFIFOClosed) {
					return
				}

				utilruntime.HandleError(err)
			}
		}
	}
}

func (e *external) RunWithContext(ctx context.Context) {
	defer utilruntime.HandleCrashWithContext(ctx)
	defer e.stopped.Store(true)

	wg := wait.Group{}
	defer wg.Wait()

	defer e.queue.Close()

	func() {
		e.startedLock.Lock()
		defer e.startedLock.Unlock()

		if e.started {
			utilruntime.HandleError(fmt.Errorf("external informer has already started - can't start it again"))
			return
		}

		e.started = true
	}()

	wg.StartWithContext(ctx, func(ctx context.Context) {
		defer utilruntime.HandleCrashWithContext(ctx)
		wait.UntilWithContext(ctx, e.processLoop, time.Second)
	})

	<-ctx.Done()
}

func (e *external) Run(stopCh <-chan struct{}) {
	e.RunWithContext(wait.ContextForChannel(stopCh))
}

func (e *external) IsStopped() bool {
	return e.stopped.Load()
}

func (e *external) Lister() cache.GenericLister {
	return cache.NewGenericLister(e.GetIndexer(), e.gvr.GroupResource())
}

func (e *external) Informer() cache.SharedIndexInformer {
	return e
}

func (e *external) HasSynced() bool {
	return e.queue.HasSynced()
}

func (e *external) SetTransform(handler cache.TransformFunc) error {
	e.startedLock.Lock()
	defer e.startedLock.Unlock()

	if e.started {
		return fmt.Errorf("external informer has already started")
	}

	e.transform = handler
	return nil
}

func (e *external) LastSyncResourceVersion() string {
	panic("external informers don't have a resource version")
}

func (e *external) SetWatchErrorHandler(_ cache.WatchErrorHandler) error {
	panic("external informers don't use kube watches")
}

func (e *external) SetWatchErrorHandlerWithContext(_ cache.WatchErrorHandlerWithContext) error {
	panic("external informers don't use kube watches")
}
