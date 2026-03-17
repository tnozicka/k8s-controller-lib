package controller

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type SyncFunc[KeyType comparable] func(ctx context.Context, key KeyType) error
type RunExtensionFunc func(ctx context.Context, wg *wait.Group)

type BaseControls[KeyType comparable] struct {
	SyncFunc         SyncFunc[KeyType]
	RunExtensionFunc RunExtensionFunc
}

type Base[KeyType comparable] struct {
	ControllerName string
	KubeClient     kubernetes.Interface
	EventRecorder  record.EventRecorder
	cachesToSync   []cache.InformerSynced
	Queue          workqueue.TypedRateLimitingInterface[KeyType]
	baseControls   BaseControls[KeyType]
}

func NewBase[KeyType comparable](
	controllerName string,
	scheme *runtime.Scheme,
	kubeClient kubernetes.Interface,
	cachesToSync []cache.InformerSynced,
	baseControls BaseControls[KeyType],
) (*Base[KeyType], error) {
	if baseControls.SyncFunc == nil {
		return nil, fmt.Errorf("SyncFunc can't be nil")
	}

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})

	return &Base[KeyType]{
		ControllerName: controllerName,
		KubeClient:     kubeClient,
		EventRecorder:  eventBroadcaster.NewRecorder(scheme, corev1.EventSource{Component: controllerName}),
		cachesToSync:   cachesToSync,
		Queue: workqueue.NewTypedRateLimitingQueueWithConfig[KeyType](
			workqueue.DefaultTypedControllerRateLimiter[KeyType](),
			workqueue.TypedRateLimitingQueueConfig[KeyType]{
				Name: controllerName,
			},
		),
		baseControls: baseControls,
	}, nil
}

func (bc *Base[KeyType]) processNextItem(ctx context.Context) bool {
	key, quit := bc.Queue.Get()
	if quit {
		return false
	}
	defer bc.Queue.Done(key)

	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()
	syncErr := bc.baseControls.SyncFunc(ctx, key)
	if syncErr == nil {
		bc.Queue.Forget(key)
		return true
	}

	switch {
	case apierrors.IsConflict(syncErr):
		klog.V(2).InfoS("Hit conflict, will retry in a bit", "Controller", bc.ControllerName, "Key", key, "Error", syncErr)

	case apierrors.IsAlreadyExists(syncErr):
		klog.V(2).InfoS("Hit already exists, will retry in a bit", "Controller", bc.ControllerName, "Key", key, "Error", syncErr)

	default:
		utilruntime.HandleError(fmt.Errorf("syncing key '%v' in controller %q failed: %w", key, bc.ControllerName, syncErr))
	}

	bc.Queue.AddRateLimited(key)

	return true
}

func (bc *Base[KeyType]) runWorker(ctx context.Context) {
	for bc.processNextItem(ctx) {
	}
}

func (bc *Base[KeyType]) Run(ctx context.Context, workers int) {
	defer utilruntime.HandleCrashWithContext(ctx)

	klog.InfoS("Starting controller", "Controller", bc.ControllerName)

	if workers <= 0 {
		utilruntime.HandleError(fmt.Errorf("at least one worker must be specified for controller %q (got %d)", bc.ControllerName, workers))
		return
	}

	var wg wait.Group
	defer func() {
		klog.InfoS("Shutting down controller", "Controller", bc.ControllerName)
		bc.Queue.ShutDown()
		wg.Wait()
		klog.InfoS("Shut down controller", "Controller", bc.ControllerName)
	}()

	if !cache.WaitForNamedCacheSync(bc.ControllerName, ctx.Done(), bc.cachesToSync...) {
		return
	}

	for range workers {
		wg.StartWithContext(ctx, func(ctx context.Context) {
			defer utilruntime.HandleCrashWithContext(ctx)

			wait.UntilWithContext(ctx, bc.runWorker, time.Second)
		})
	}

	if bc.baseControls.RunExtensionFunc != nil {
		bc.baseControls.RunExtensionFunc(ctx, &wg)
	}

	<-ctx.Done()
}
