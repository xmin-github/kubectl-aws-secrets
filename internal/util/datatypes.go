package util

//SecretConfig is secret data structur
type SecretConfig struct {
	Secret  string
	Sources map[string][]string
}

//KeyPair is keypair structure
type KeyPair struct {
	Key   string
	Value string
}
