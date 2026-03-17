package cacmreconciler

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"

	"github.com/tnozicka/k8s-controller-lib/pkg/genericclioptions"
)

type Options struct {
	*genericclioptions.ClientConfig
	*genericclioptions.InClusterReflection
	*genericclioptions.LeaderElection

	MyArg string

	kubeClient kubernetes.Interface
}

func NewOptions(_ genericclioptions.IOStreams, programName string) *Options {
	return &Options{
		ClientConfig:        genericclioptions.NewClientConfig(programName),
		InClusterReflection: genericclioptions.NewInClusterReflection(),
		LeaderElection:      genericclioptions.NewLeaderElection(programName),

		MyArg: "MyArgDefaultValue",
	}
}

func (o *Options) AddFlags(cmd *cobra.Command) {
	o.ClientConfig.AddFlags(cmd.Flags())
	o.InClusterReflection.AddFlags(cmd.Flags())
	o.LeaderElection.AddFlags(cmd.Flags())

	cmd.Flags().StringVarP(&o.MyArg, "my-flag", "", o.MyArg, "MyArg description")
}

func (o *Options) Validate(_ []string) error {
	var errs []error

	errs = append(errs, o.ClientConfig.Validate())
	errs = append(errs, o.InClusterReflection.Validate())
	errs = append(errs, o.LeaderElection.Validate())

	if len(o.MyArg) == 0 {
		errs = append(errs, errors.New("my-flag can't be empty"))
	}

	return errors.Join(errs...)
}

func (o *Options) Complete(_ []string) error {
	err := o.ClientConfig.Complete()
	if err != nil {
		return err
	}

	err = o.InClusterReflection.Complete()
	if err != nil {
		return err
	}

	err = o.LeaderElection.Complete()
	if err != nil {
		return err
	}

	o.kubeClient, err = kubernetes.NewForConfig(o.ProtoConfig)
	if err != nil {
		return fmt.Errorf("can't create kubernetes clientset: %w", err)
	}

	return nil
}
