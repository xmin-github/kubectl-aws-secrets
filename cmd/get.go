package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xmin-github/kubectl-secrets/internal/awssm"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func createGetCommand(streams genericclioptions.IOStreams) *cobra.Command {
	secret := &awsSecret{
		out: streams.Out,
	}

	sCmd := &cobra.Command{
		Use:          "get",
		Short:        "import secrets from AWS Secrets Manager",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			secretName, err := c.Flags().GetString("aws-secret-id")
			if err != nil {
				return err
			}

			/*if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}*/
			return secret.getCmdExecute(secretName)
		},
	}
	return sCmd
}

func (sv *awsSecret) getCmdExecute(secretName string) error {
	secret, err := awssm.GetSecret(secretName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(sv.out, "AWS Secrets Manager item: %s %s\n", secretName, secret)
	if err != nil {
		return err
	}

	return nil
}

var getSecretCmd = createGetCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

func init() {
	getSecretCmd.Flags().StringVarP(&awsSecretID, "aws-secret-id", "a", "", "Secret Name in AWS Secrets Manager")
	getSecretCmd.MarkFlagRequired("aws-secret-id")
	rootCmd.AddCommand(getSecretCmd)
}
