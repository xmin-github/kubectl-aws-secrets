package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xmin-github/kubectl-aws-secrets/internal/awssm"
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
			secretName, err := c.Flags().GetString("aws-key-name")
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

	fmt.Printf("IsJSONString(%s) = %v\n", secret, IsJSONString(secret))
	fmt.Printf("IsJSON(%s) = %v\n", secret, IsJSON(secret))
	if IsJSON(secret) {
		sKey, sValue, err := ParseKeyPair(secret)
		if err != nil {
			return err
		}

		fmt.Fprintf(sv.out, "AWS secret key: %s, value: %s\n", sKey, sValue)
	}

	return nil
}

var getSecretCmd = createGetCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

func init() {
	getSecretCmd.Flags().StringVarP(&awsKeyName, "aws-key-name", "a", "", "Secret Name in AWS Secrets Manager")
	getSecretCmd.Flags().StringVarP(&iamRole, "role-arn", "r", "", "aws iam role")
	getSecretCmd.MarkFlagRequired("aws-key-name")
	rootCmd.AddCommand(getSecretCmd)
}
