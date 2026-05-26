package cacmreconciler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server/healthz"
	kubeinformers "k8s.io/client-go/informers"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/examples/pkg/controller/cacmreconciler"
	"github.com/tnozicka/k8s-controller-lib/examples/pkg/scheme"
	"github.com/tnozicka/k8s-controller-lib/pkg/genericclioptions"
	clhealthz "github.com/tnozicka/k8s-controller-lib/pkg/healthz"
	clhttp "github.com/tnozicka/k8s-controller-lib/pkg/http"
	"github.com/tnozicka/k8s-controller-lib/pkg/leaderelection"
	"github.com/tnozicka/k8s-controller-lib/pkg/version"
)

const (
	resyncPeriod = 24 * time.Hour
)

func (o *Options) Run(ctx context.Context, _ genericclioptions.IOStreams, cmd *cobra.Command) error {
	var wg wait.Group
	defer wg.Wait()

	ctx, ctxCancel := context.WithCancelCause(ctx)
	defer ctxCancel(nil)

	klog.InfoS("Starting", "Command", cmd.Name(), "version", version.Get())
	cliflag.PrintFlags(cmd.Flags())

	kubeInformers := kubeinformers.NewSharedInformerFactory(
		o.kubeClient,
		resyncPeriod,
	)

	controller, err := cacmreconciler.NewController(
		scheme.Scheme,
		"internal.k8s-controller-lib/managed-hash", // Use your own, including a full domain name.
		o.kubeClient,
		kubeInformers.Core().V1().Secrets(),
		kubeInformers.Core().V1().ConfigMaps(),
	)
	if err != nil {
		return fmt.Errorf("can't create agent controller: %w", err)
	}

	mux := http.NewServeMux()
	healthz.InstallReadyzHandler(
		mux,
		healthz.NewShutdownHealthz(ctx.Done()),
		healthz.NamedCheck("informer-sync", func(req *http.Request) error {
			return clhealthz.HealthCheckers{
				healthz.NewInformerSyncHealthz(kubeInformers),
			}.Check(req)
		}),
		o.LeaderElection.WatchDog,
	)
	healthz.InstallHandler(mux, healthz.PingHealthz)

	server := clhttp.NewServer(&http.Server{
		Addr:    ":5000",
		Handler: mux,
	})

	kubeInformers.Start(ctx.Done())
	defer kubeInformers.Shutdown()

	// Start probe server.
	wg.StartWithContext(ctx, func(ctx context.Context) {
		lsErr := server.ListenAndServe(ctx)
		if lsErr != nil {
			runtime.HandleError(fmt.Errorf("probe server failed: %w", lsErr))
		}
	})

	identity, err := leaderelection.MakeHostnameIdentity()
	if err != nil {
		return fmt.Errorf("can't make hostname identity: %w", err)
	}

	return o.LeaderElection.Run(
		ctx,
		cmd.Name(),
		types.NamespacedName{
			Namespace: o.Namespace,
			Name:      corev1.NamespaceDefault, // Adjust it to your namespace name.
		},
		o.kubeClient,
		identity,
		func(ctx context.Context) error {
			wg.StartWithContext(ctx, func(ctx context.Context) {
				controller.Run(ctx, 1)
			})

			<-ctx.Done()
			return nil
		},
	)
}
