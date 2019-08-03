package k8sutil

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

//CreateSecret creates a Secret object in Kubernetes
func CreateSecret(secretName string, dataKey string, dataValue string, forceUpdate bool) (string, error) {
	var kubeconfig *string
	var configEnv string
	fmt.Printf("force update %v\n", forceUpdate)
	configEnv = os.Getenv("KUBECONFIG")
	if configEnv != "" {
		kubeconfig = &configEnv
	} else {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
		flag.Parse()
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	secret := &apiv1.Secret{
		/*TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},*/
		Type: apiv1.SecretTypeOpaque,
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			dataKey: []byte(dataValue),
		},
	}
	secretClient := clientset.CoreV1().Secrets(apiv1.NamespaceDefault)

	result, err := secretClient.Create(secret)
	if err == nil {
		fmt.Printf("%q created \n", result.GetObjectMeta().GetName())
		return result.GetObjectMeta().GetName(), nil
	} else if strings.Contains(err.Error(), "already exists") && forceUpdate {
		//update the secrete
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of Deployment before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			currSecret, getErr := secretClient.Get(secretName, metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get secret %s: %v", secretName, getErr))
			}
			if currSecret.Data != nil {
				currSecret.Data[dataKey] = []byte(dataValue)
			} else {
				currSecret.Data = make(map[string][]byte)
				currSecret.Data[dataKey] = []byte(dataValue)
			}
			_, updateErr := secretClient.Update(currSecret)
			if updateErr == nil {
				fmt.Printf("secret %s is updated\n", secretName)
			}
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update secrete %s failed: %v", secretName, retryErr))
		}

	} else {
		return secretName, err

	}
	return secretName, nil
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func int32Ptr(i int32) *int32 { return &i }
