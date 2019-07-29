package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xmin-github/kubectl-aws-secrets/internal/awssm"
	"github.com/xmin-github/kubectl-aws-secrets/internal/k8sutil"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func createImportCommand(streams genericclioptions.IOStreams) *cobra.Command {
	secret := &awsSecret{
		out: streams.Out,
	}

	sCmd := &cobra.Command{
		Use:          "import",
		Short:        "import secrets from AWS Secrets Manager",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			secretName, err := c.Flags().GetString("aws-secret-id")
			if err != nil {
				return err
			}
			k8sSecretName, err := c.Flags().GetString("k8s-secret-name")
			if err != nil {
				return err
			}
			forceUpdate, err := c.Flags().GetBool("force")
			if err != nil {
				return err
			}

			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}
			return secret.importCmdExecute(k8sSecretName, secretName, forceUpdate)
		},
	}
	return sCmd
}

func (sv *awsSecret) importCmdExecute(k8sSecretName string, secretName string, forceUpdate bool) error {

	secretValue, err := awssm.GetSecret(secretName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(sv.out, "AWS Secrets Manager item: %s %s\n", secretName, secretValue)
	if err != nil {
		return err
	}

	result, err := k8sutil.CreateSecret(k8sSecretName, secretName, secretValue, forceUpdate)
	if err != nil {
		fmt.Fprintf(sv.out, "Kubernete Secret: %s\n", err.Error())
	} else {
		fmt.Fprintf(sv.out, "Kubernete Secret: %s\n", result)
	}

	return nil
}

var importSecretCmd = createImportCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

func init() {
	importSecretCmd.Flags().StringVarP(&awsSecretID, "aws-secret-id", "a", "", "Secret Name in AWS Secrets Manager")
	importSecretCmd.Flags().StringVarP(&k8sSecretName, "k8s-secret-name", "k", "", "Secret object name in ks")
	importSecretCmd.Flags().BoolVarP(&force, "force", "f", false, "if a secret exists, update it ")
	importSecretCmd.MarkFlagRequired("aws-secret-id")
	importSecretCmd.MarkFlagRequired("k8s-secret-name")
	rootCmd.AddCommand(importSecretCmd)
}
