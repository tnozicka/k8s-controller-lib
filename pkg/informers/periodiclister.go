package informers

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"github.com/tnozicka/k8s-controller-lib/pkg/helpers/slices"
)

type Lister interface {
	List(ctx context.Context) ([]runtime.Object, error)
}

type ListFunc func(ctx context.Context) ([]runtime.Object, error)

var _ Lister = new(ListFunc)

func (l ListFunc) List(ctx context.Context) ([]runtime.Object, error) {
	return (l)(ctx)
}

type periodicLister struct {
	External
	stopped atomic.Bool

	lister   Lister
	interval time.Duration
}

var _ cache.SharedIndexInformer = &periodicLister{}
var _ informers.GenericInformer = &periodicLister{}

func NewPeriodicLister(gvr schema.GroupVersionResource, keyFunc cache.KeyFunc, lister Lister, interval time.Duration) informers.GenericInformer {
	return &periodicLister{
		External: NewExternal(gvr, keyFunc, nil),
		lister:   lister,
		interval: interval,
	}
}

func (p *periodicLister) list(ctx context.Context) error {
	klog.V(6).InfoS("Listing objects")

	objs, err := p.lister.List(ctx)
	if err != nil {
		return fmt.Errorf("can't list objects: %w", err)
	}

	err = p.ReplaceObjects(
		slices.Convert(
			func(from runtime.Object) any {
				return from
			},
			objs...,
		),
	)
	if err != nil {
		return fmt.Errorf("can't replace objects: %w", err)
	}

	return nil
}

func (p *periodicLister) runLister(ctx context.Context) {
	klog.V(4).InfoS("Starting periodic lister", "GVR", p.External.Name(), "Interval", p.interval)
	defer klog.V(4).InfoS("Periodic lister finished", "GVR", p.External.Name())

	t := time.NewTimer(p.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		t.Reset(p.interval)

		err := p.list(ctx)
		if err != nil {
			klog.ErrorS(err, "list failed")
			return
		}

		select {
		case <-t.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (p *periodicLister) RunWithContext(ctx context.Context) {
	defer utilruntime.HandleCrashWithContext(ctx)
	defer p.stopped.Store(true)

	var wg wait.Group
	defer wg.Wait()

	wg.StartWithContext(ctx, p.External.RunWithContext)

	delayHandler := wait.Backoff{
		Duration: 800 * time.Millisecond,
		Factor:   2.0,
		Jitter:   0.1,
		Cap:      2 * time.Minute,
	}.DelayWithReset(&clock.RealClock{}, 30*time.Second)
	// Currently, the kube machinery is broken and we can't use a deprecated function.
	// Looks like someone forgot to implement BackoffManager interface in a non-deprecated function,
	// so we can't use BackoffUntil.
	_ = delayHandler.Until(ctx, true, true, func(fCtx context.Context) (bool, error) {
		p.runLister(fCtx)
		return false, nil
	})
}

func (p *periodicLister) Run(stopCh <-chan struct{}) {
	p.RunWithContext(wait.ContextForChannel(stopCh))
}

func (p *periodicLister) Informer() cache.SharedIndexInformer {
	return p
}

func (p *periodicLister) IsStopped() bool {
	return p.External.IsStopped() && p.stopped.Load()
}

func (p *periodicLister) GetController() cache.Controller {
	return &controllerAdapter{
		lastSyncResourceVersion: p.External.LastSyncResourceVersion,
		hasSynced:               p.External.HasSynced,
		run:                     p.RunWithContext,
	}
}
