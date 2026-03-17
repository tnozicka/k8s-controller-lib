package cmdutil

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
)

const (
	FlagLogLevelKey = "loglevel"
)

func NormalizeNameForEnvVar(name string) string {
	s := strings.ToUpper(name)
	s = strings.ReplaceAll(s, "-", "_")
	return s
}

func ReadFlagsFromEnv(prefix string, cmd *cobra.Command) error {
	var errs []error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			// flags always take precedence over environment
			return
		}

		// See if there exists matching environment variable
		envVarName := NormalizeNameForEnvVar(prefix + f.Name)
		v, exists := os.LookupEnv(envVarName)
		if !exists {
			return
		}

		err := f.Value.Set(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("can't parse env var %q with value %q into flag %q: %v", envVarName, v, f.Name, err))
			return
		}

		f.Changed = true
	})

	return errors.Join(errs...)
}

func InstallKlog(cmd *cobra.Command) error {
	const klogVFlagName = "v"
	vFlag := flag.CommandLine.Lookup(klogVFlagName)
	if vFlag == nil {
		return fmt.Errorf("can't lookup klog %q flag", klogVFlagName)
	}
	level := vFlag.Value.(*klog.Level)
	levelPtr := (*int32)(level)
	cmd.PersistentFlags().Int32Var(levelPtr, FlagLogLevelKey, *levelPtr, "Set the level of log output (0-10).")
	if cmd.PersistentFlags().Lookup(klogVFlagName) == nil {
		cmd.PersistentFlags().Int32Var(levelPtr, klogVFlagName, *levelPtr, "Set the level of log output (0-10).")
	}
	cmd.PersistentFlags().Lookup(klogVFlagName).Hidden = true

	// Enable directory prefix.
	const klogAddDirHeaderName = "add_dir_header"
	addDirHeaderFlag := flag.CommandLine.Lookup(klogAddDirHeaderName)
	if addDirHeaderFlag == nil {
		return fmt.Errorf("can't lookup klog %q flag", klogAddDirHeaderName)
	}
	addDirHeaderValue := "true"
	err := addDirHeaderFlag.Value.Set(addDirHeaderValue)
	if err != nil {
		return fmt.Errorf("can't set klog %q flag to %q", klogAddDirHeaderName, addDirHeaderValue)
	}

	return nil
}

func GetLoglevel() string {
	f := flag.CommandLine.Lookup("v")
	if f != nil {
		return f.Value.String()
	}

	return ""
}
