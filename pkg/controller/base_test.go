package controller

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"
)

func TestNewBase(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name                    string
		cachesToSync            []cache.InformerSynced
		syncFunc                SyncFunc[string]
		expectedControllerName  string
		expectedCachesToSyncLen int
		expectedErr             error
	}{
		{
			name:                    "nil SyncFunc",
			cachesToSync:            nil,
			syncFunc:                nil,
			expectedControllerName:  "",
			expectedCachesToSyncLen: 0,
			expectedErr:             fmt.Errorf("SyncFunc can't be nil"),
		},
		{
			name:                    "valid SyncFunc",
			cachesToSync:            []cache.InformerSynced{func() bool { return true }},
			syncFunc:                func(_ context.Context, _ string) error { return nil },
			expectedControllerName:  "test-controller",
			expectedCachesToSyncLen: 1,
			expectedErr:             nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			scheme := runtime.NewScheme()
			err := corev1.AddToScheme(scheme)
			if err != nil {
				t.Fatalf("can't add corev1 to scheme: %v", err)
			}

			kubeClient := fake.NewClientset()
			base, err := NewBase[string](
				"test-controller",
				scheme,
				kubeClient,
				tc.cachesToSync,
				BaseControls[string]{
					SyncFunc: tc.syncFunc,
				},
			)

			if !reflect.DeepEqual(err, tc.expectedErr) {
				t.Errorf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, err))
			}

			if base == nil {
				return
			}

			if base.ControllerName != tc.expectedControllerName {
				t.Errorf("expected and got ControllerName differ:\n%s",
					cmp.Diff(tc.expectedControllerName, base.ControllerName))
			}

			if len(base.cachesToSync) != tc.expectedCachesToSyncLen {
				t.Errorf("expected cachesToSync length %d, got %d",
					tc.expectedCachesToSyncLen, len(base.cachesToSync))
			}

			if base.KubeClient == nil {
				t.Error("expected KubeClient to be set, got nil")
			}

			if base.Queue == nil {
				t.Error("expected Queue to be set, got nil")
			}
		})
	}
}

func TestBase_Run_gracefulTermination(t *testing.T) {
	t.Parallel()

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("can't add corev1 to scheme: %v", err)
	}

	tt := []struct {
		name             string
		cachesToSync     []cache.InformerSynced
		expectedFinished bool
	}{
		{
			name: "gracefully terminates when caches are synced",
			cachesToSync: []cache.InformerSynced{
				func() bool {
					return true
				},
			},
			expectedFinished: true,
		},

		{
			name: "gracefully terminates when caches aren't synced",
			cachesToSync: []cache.InformerSynced{
				func() bool {
					return false
				},
			},
			expectedFinished: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cacheSyncCalledChan := make(chan struct{})
			cachesToSync := make([]cache.InformerSynced, 0, 1+len(tc.cachesToSync))
			cachesToSync = append(cachesToSync, func() bool {
				close(cacheSyncCalledChan)
				return true
			})
			cachesToSync = append(cachesToSync, tc.cachesToSync...)

			kubeClient := fake.NewClientset()
			base, err := NewBase[string](
				"test-run-controller",
				scheme,
				kubeClient,
				cachesToSync,
				BaseControls[string]{
					SyncFunc: func(_ context.Context, _ string) error { return nil },
				},
			)
			if err != nil {
				t.Fatalf("can't create Base: %v", err)
			}

			ctx, cancelCtx := context.WithCancel(t.Context())
			defer cancelCtx()

			var wg wait.Group
			defer wg.Wait()

			runFinishedChan := make(chan struct{})
			wg.StartWithContext(ctx, func(ctx context.Context) {
				base.Run(ctx, 1)
				close(runFinishedChan)
			})

			const cacheCallTimeout = 5 * time.Second
			select {
			case <-cacheSyncCalledChan:
				break
			case <-time.After(cacheCallTimeout):
				t.Fatalf("cache sync method didn't get called within %v", cacheCallTimeout)
			}

			cancelCtx()

			const gracefulTerminationTimeout = 5 * time.Second
			select {
			case <-runFinishedChan:
				if !tc.expectedFinished {
					t.Error("Run finished unexpectedly")
				}
			case <-time.After(gracefulTerminationTimeout):
				t.Errorf("Run didn't finish within graceful termination period of %v", gracefulTerminationTimeout)
			}
		})
	}
}
