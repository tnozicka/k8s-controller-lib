package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

type server struct {
	*http.Server

	resolvedListenAddr   string
	resolvedListenAddrCh chan struct{}
}

type Server interface {
	ListenAndServe(ctx context.Context) error
}

func NewServer(httpServer *http.Server) Server {
	return &server{
		Server:               httpServer,
		resolvedListenAddrCh: make(chan struct{}),
	}
}

func (s *server) ListenAndServe(ctx context.Context) error {
	defer utilruntime.HandleCrashWithContext(ctx)

	ctx, ctxCancel := context.WithCancel(ctx)
	defer ctxCancel()

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return fmt.Errorf("can't create listener: %w", err)
	}
	defer func() {
		closeErr := listener.Close()
		if closeErr != nil && !errors.Is(closeErr, net.ErrClosed) {
			klog.ErrorS(closeErr, "Error closing listener")
		}
	}()

	s.resolvedListenAddr = listener.Addr().String()
	close(s.resolvedListenAddrCh)

	finishedServing := make(chan struct{})

	var wg wait.Group
	defer wg.Wait()

	var shutdownErr error
	wg.StartWithContext(ctx, func(ctx context.Context) {
		select {
		case <-ctx.Done():
			shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 12*time.Second)
			defer shutdownCtxCancel()

			klog.V(2).InfoS("Shutting down HTTP server", "Address", s.Addr)
			shutdownErr = s.Server.Shutdown(shutdownCtx)
			return

		case <-finishedServing:
			shutdownErr = nil
			return
		}
	})

	var listenErr error
	wg.StartWithContext(ctx, func(_ context.Context) {
		klog.V(2).InfoS("Starting HTTP server", "Address", s.resolvedListenAddr)
		defer klog.V(2).InfoS("HTTP server done serving", "Address", s.resolvedListenAddr)

		listenErr = s.Server.Serve(listener)
		close(finishedServing)
		if errors.Is(listenErr, http.ErrServerClosed) {
			listenErr = nil
		}
	})

	// Wait here explicitly, so the errors make it through.
	wg.Wait()

	return errors.Join(listenErr, shutdownErr)
}
