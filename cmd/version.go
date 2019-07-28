package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const appLabel = "kubectl aws-secrets version"

var version = "0.0.1"

type versionCmd struct {
	out io.Writer
}

func newVersionCmd(streams genericclioptions.IOStreams) *cobra.Command {
	version := &versionCmd{
		out: streams.Out,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "print the version number and exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return version.run()
		},
	}
	return cmd
}

func (v *versionCmd) run() error {
	_, err := fmt.Fprintf(v.out, "%s %s\n", appLabel, version)
	if err != nil {
		return err
	}
	return nil
}

var theVersionCmd = newVersionCmd(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

func init() {
	rootCmd.AddCommand(theVersionCmd)
}
