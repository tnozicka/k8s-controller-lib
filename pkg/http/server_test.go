package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/clock"
)

func TestServer_ListenAndServe(t *testing.T) {
	t.Parallel()

	newDelay := func() wait.DelayFunc {
		return wait.Backoff{
			Duration: 800 * time.Millisecond,
			Factor:   2.0,
			Jitter:   0.1,
			Cap:      2 * time.Minute,
		}.DelayWithReset(&clock.RealClock{}, 30*time.Second)
	}

	type srvFunc func(t *testing.T) (*server, context.Context, func())
	tt := []struct {
		name        string
		srvFunc     srvFunc
		expectedErr error
	}{
		{
			name: "graceful shutdown on context cancellation while running",
			srvFunc: func(t *testing.T) (*server, context.Context, func()) {
				t.Helper()

				ready := make(chan struct{})
				s := NewServer(&http.Server{
					Addr: "localhost:",
					Handler: http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
						close(ready)
					}),
				}).(*server)
				ctx, ctxCancel := context.WithCancel(t.Context())
				return s,
					ctx,
					func() {
						waitErr := newDelay().Until(t.Context(), true, true, func(_ context.Context) (bool, error) {
							<-s.resolvedListenAddrCh
							_, err := http.Get("http://" + s.resolvedListenAddr)
							if err != nil {
								t.Logf("HTTP server is not ready yet: %v", err)
								return false, err
							}

							t.Log("HTTP server is ready")
							ctxCancel()
							return true, nil

						})
						if waitErr != nil {
							t.Fatalf("can't wait for server to be ready: %v", waitErr)
						}
					}
			},
			expectedErr: nil,
		},
		{
			name: "already canceled context",
			srvFunc: func(t *testing.T) (*server, context.Context, func()) {
				ctx, ctxCancel := context.WithCancelCause(t.Context())
				ctxCancel(errors.New("always canceled test context"))
				return NewServer(&http.Server{
						Addr: ":",
					}).(*server),
					ctx,
					func() {}
			},
			expectedErr: nil,
		},
		{
			name: "listen error on invalid address",
			srvFunc: func(t *testing.T) (*server, context.Context, func()) {
				return NewServer(&http.Server{
						Addr: ":-42",
					}).(*server),
					t.Context(),
					func() {}
			},
			expectedErr: fmt.Errorf("can't create listener: %w",
				&net.OpError{
					Op:  "listen",
					Net: "tcp",
					Err: &net.AddrError{
						Err:  "invalid port",
						Addr: "-42",
					},
				},
			),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s, ctx, waitFunc := tc.srvFunc(t)

			var wg wait.Group

			var gotErr error
			wg.Start(func() {
				gotErr = s.ListenAndServe(ctx)
				t.Logf("Listen finished with err %v", gotErr)
			})

			wg.Start(waitFunc)

			wg.Wait()

			if !reflect.DeepEqual(gotErr, tc.expectedErr) {
				t.Errorf("expected and got errors differ:\n%s", cmp.Diff(tc.expectedErr, gotErr, cmpopts.EquateErrors()))
			}
		})
	}
}
