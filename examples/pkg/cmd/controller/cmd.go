package controller

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/tnozicka/k8s-controller-lib/examples/pkg/cmd/controller/cacmreconciler"
	"github.com/tnozicka/k8s-controller-lib/pkg/cmdutil"
	"github.com/tnozicka/k8s-controller-lib/pkg/genericclioptions"
)

const (
	EnvVarPrefix = "CONTROLLER_"
)

type Options struct {
}

func NewOptions(_ genericclioptions.IOStreams) *Options {
	return &Options{}
}

func (o *Options) AddPersistentFlags(_ *pflag.FlagSet) {
}

func NewCommand(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	programName := "controller"
	cmd := &cobra.Command{
		Use: programName,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			err := cmdutil.ReadFlagsFromEnv(EnvVarPrefix, cmd)
			if err != nil {
				return fmt.Errorf("can't read flags from env: %w", err)
			}

			return nil
		},
	}

	runtime.Must(cmdutil.InstallKlog(cmd))

	cmd.AddCommand(cacmreconciler.NewCommand(streams, programName))

	o.AddPersistentFlags(cmd.PersistentFlags())

	return cmd
}
