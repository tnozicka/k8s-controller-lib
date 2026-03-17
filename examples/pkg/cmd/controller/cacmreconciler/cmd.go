package cacmreconciler

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/tnozicka/k8s-controller-lib/pkg/genericclioptions"
)

func NewCommand(streams genericclioptions.IOStreams, programName string) *cobra.Command {
	o := NewOptions(streams, programName)

	cmd := &cobra.Command{
		Use: "run-ca-cm-reconciler",
		Long: templates.LongDesc(`
		`),
		ValidArgs: nil,
		Short:     "Runs CA ConfigMap reconciler controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.Validate(args)
			if err != nil {
				return err
			}

			err = o.Complete(args)
			if err != nil {
				return err
			}

			err = o.Run(cmd.Context(), streams, cmd)
			if err != nil {
				return err
			}

			return nil
		},

		SilenceErrors: true,
		SilenceUsage:  true,
	}

	o.AddFlags(cmd)

	return cmd
}
