package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string
var awsSecretID string
var k8sSecretName string
var force bool

type awsSecret struct {
	out io.Writer
}

var rootCmd = &cobra.Command{
	Use:   "aws-secrets [command] [flags]",
	Short: "import secret from aws",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at http://hugo.spf13.com`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

//Execute the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {

}

//IsJSONString checks if a string a quoted string
func IsJSONString(s string) bool {
	var js string
	return json.Unmarshal([]byte(s), &js) == nil

}

//IsJSON checks if a string is a json object
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

//ParseKeyPair gets key and value from a key name/value pair json object
//Assumption: the json contains only one key/value pair
func ParseKeyPair(secret string) (key string, value string, err error) {
	j := []byte(secret)
	c := make(map[string]interface{})
	e := json.Unmarshal(j, &c)
	if e != nil {
		return "", "", e
	}
	for key, value := range c {
		if valueStr, ok := value.(string); ok {
			return key, valueStr, nil
		}
	}
	unKnownErr := fmt.Errorf("key and value parsing from aws secret json failed")
	return "", "", unKnownErr

}
