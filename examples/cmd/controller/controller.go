package main

import (
	"flag"
	"os"

	"k8s.io/klog/v2"

	"github.com/tnozicka/k8s-controller-lib/examples/pkg/cmd/controller"
	"github.com/tnozicka/k8s-controller-lib/pkg/genericclioptions"
	"github.com/tnozicka/k8s-controller-lib/pkg/leaderelection"
	"github.com/tnozicka/k8s-controller-lib/pkg/signals"
)

func exit(code int) {
	flushLogs()
	os.Exit(code) // nolint:gocritic
}

func flushLogs() {
	klog.InfoS("Flushing logs")
	klog.Flush()
}

func main() {
	klog.InitFlags(flag.CommandLine)
	err := flag.Set("logtostderr", "false")
	if err != nil {
		panic(err)
	}
	err = flag.Set("alsologtostderr", "true")
	if err != nil {
		panic(err)
	}

	ctx := signals.GetContext()

	leaderelection.ElectionLostCallback = func() {
		flushLogs()
	}

	command := controller.NewCommand(genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	})
	err = command.ExecuteContext(ctx)
	if err != nil {
		klog.ErrorS(err, "Command execution failed", "CommandName", command.Name())
		exit(1)
	} else {
		exit(0)
	}

}
