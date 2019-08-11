package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/xmin-github/kubectl-aws-secrets/internal/awssm"
	"github.com/xmin-github/kubectl-aws-secrets/internal/k8sutil"
	"github.com/xmin-github/kubectl-aws-secrets/internal/util"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func readSecretFile(filePath string) (*util.SecretConfig, error) {
	var sConfig util.SecretConfig
	source, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(source, &sConfig)
	if err != nil {
		return nil, err
	}

	return &sConfig, nil
}

func getSecretKeys(source string, sConfig *util.SecretConfig) ([]string, bool) {
	src, ok := sConfig.Sources[source]
	return src, ok

}

func validateOptions(awsKeyName string, source string, file string) error {
	if awsKeyName != "" && file != "" {
		return fmt.Errorf("--aws-key-name and --file cannot be used at the same time")
	} else if awsKeyName != "" && source == "" {
		return fmt.Errorf("--source is required while --aws-key-name is used")
	} else if awsKeyName == "" && file == "" {
		return fmt.Errorf("--aws-key-name or --file is required")
	} else {
		return nil
	}

}
func createImportCommand(streams genericclioptions.IOStreams) *cobra.Command {
	secret := &awsSecret{
		out: streams.Out,
	}

	sCmd := &cobra.Command{
		Use:          "import",
		Short:        "import AWS Secrets or SSM Parameters as k8s Secret",
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			awsKeyName, err := c.Flags().GetString("aws-key-name")
			if err != nil {
				return err
			}

			source, err := c.Flags().GetString("source")
			if err != nil {
				return err
			}

			filePath, err := c.Flags().GetString("file")
			if err != nil {
				return err
			}
			validationErr := validateOptions(awsKeyName, source, filePath)
			if validationErr != nil {
				return validationErr
			}
			ns, err := c.Flags().GetString("namespace")
			if err != nil {
				return err
			}
			k8sSecretName, err := c.Flags().GetString("k8s-secret-name")
			if err != nil {
				return err
			}

			update, err := c.Flags().GetBool("update")
			if err != nil {
				return err
			}

			if len(args) != 0 {
				return errors.New("this command does not accept arguments")
			}

			var execErr error
			if awsKeyName != "" {
				if source == "aws-secrets" {
					execErr = secret.importCmdSecretExecute(k8sSecretName, awsKeyName, ns, update)
				} else if source == "aws-ssm" {
					execErr = secret.importCmdParamExecute(k8sSecretName, awsKeyName, ns, update)
				} else {
					execErr = fmt.Errorf("error: source is not passed")
				}
			} else if filePath != "" {
				sConfig, err := readSecretFile(filePath)
				if err != nil {
					panic(err)

				} else {

					execErr = secret.importCmdBatchExecute(sConfig, ns, update)
				}

			} else {
				execErr = fmt.Errorf("--aws-key-name or --file is required")
			}
			return execErr
		},
	}
	return sCmd
}

func (sv *awsSecret) importCmdSecretExecute(k8sSecretName string, secretName string, ns string, update bool) error {
	dataKey, dataValue, err := getSecret(secretName)
	if err != nil {
		return err
	}
	if k8sSecretName == "" {
		k8sSecretName = secretName
	}
	result, err := k8sutil.CreateSecret(k8sSecretName, dataKey, dataValue, ns, update)
	if err != nil {
		fmt.Fprintf(sv.out, "Kubernete Secret: %s\n", err.Error())
		return err
	}
	fmt.Fprintf(sv.out, "Kubernete Secret: %s\n", result)
	return nil
}

func (sv *awsSecret) importCmdParamExecute(k8sSecretName string, paramName string, ns string, update bool) error {

	secret, err := awssm.GetParameter(paramName)
	var dataKey, dataValue string
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(sv.out, "AWS SSM Parameter: %s\n", paramName)
	if err != nil {
		return err
	}

	if k8sSecretName == "" {
		k8sSecretName = paramName
	}
	dataKey = paramName
	dataValue = secret
	fmt.Printf("paramValue: %s\n", dataValue)
	result, err := k8sutil.CreateSecret(k8sSecretName, dataKey, dataValue, ns, update)
	if err != nil {
		fmt.Fprintf(sv.out, "Kubernete Secret: %s\n", err.Error())
	} else {
		fmt.Fprintf(sv.out, "Kubernete Secret: %s\n", result)
	}

	return nil
}

func getSecret(secretName string) (key string, value string, e error) {
	secret, err := awssm.GetSecret(secretName)
	var dataKey, dataValue string
	if err != nil {
		return dataKey, dataValue, err
	}

	if IsJSON(secret) {
		dataKey, dataValue, err = ParseKeyPair(secret)
		if err != nil {
			return dataKey, dataValue, err
		}
		return dataKey, dataValue, nil
	}
	dataKey = secretName
	dataValue = secret
	return dataKey, dataValue, nil
}

func (sv *awsSecret) importCmdBatchExecute(sConfig *util.SecretConfig, ns string, update bool) error {
	var secretKeys []string
	var found bool
	var secretDataItems []util.KeyPair = []util.KeyPair{}

	secretKeys, found = getSecretKeys("aws-secrets", sConfig)
	if found {
		fmt.Printf("aws-secrets size: %d\n", len(secretKeys))
		for _, keyName := range secretKeys {
			dataKey, dataValue, err := getSecret(keyName)
			if err == nil {
				secretDataItems = append(secretDataItems, util.KeyPair{Key: dataKey, Value: dataValue})
			} else {
				fmt.Printf("aws-secrets %s not found\n", keyName)
			}
		}
	}

	secretKeys, found = getSecretKeys("aws-ssm", sConfig)
	if found {
		fmt.Printf("aws-secrets size: %d\n", len(secretKeys))
		for _, paramName := range secretKeys {
			dataValue, err := awssm.GetParameter(paramName)
			if err == nil {
				secretDataItems = append(secretDataItems, util.KeyPair{Key: paramName, Value: dataValue})
			} else {
				fmt.Printf("aws-ssm %s not found\n", paramName)
			}
		}
	}
	for _, item := range secretDataItems {
		fmt.Printf("key: %s, value: %s\n", item.Key, "xxxxx")
	}
	result, err := k8sutil.CreateSecretBatch(sConfig.Secret, secretDataItems, ns, update)
	if err != nil {
		return err
	}
	fmt.Printf("k6s secret %s is created/updated\n", result)
	return nil
}

var importSecretCmd = createImportCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

func init() {
	importSecretCmd.Flags().StringVarP(&awsKeyName, "aws-key-name", "a", "", "AWS Secret/Parameter Name in AWS Secrets Manager")
	importSecretCmd.Flags().StringVarP(&source, "source", "s", "", "Valid values: aws-secrets, or aws-ssm")
	importSecretCmd.Flags().StringVarP(&k8sSecretName, "k8s-secret-name", "k", "", "Secret object name in ks")
	importSecretCmd.Flags().StringVarP(&nameSpace, "namespace", "n", "default", "Namespace for the secret")
	importSecretCmd.Flags().BoolVarP(&update, "update", "u", false, "if a secret exists, update it ")
	//importSecretCmd.Flags().StringVarP(&iamRole, "role-arn", "r", "", "aws iam role")
	importSecretCmd.Flags().StringVarP(&secretFilePath, "file", "f", "", "secret config filepath")
	//importSecretCmd.MarkFlagRequired("aws-secret-id")
	rootCmd.AddCommand(importSecretCmd)
}
