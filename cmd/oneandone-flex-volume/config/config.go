package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/glog"
)

const (
	tokenFileEnv         = "ONEANDONE_TOKEN_FILE_PATH"
	tokenEnv             = "ONEANDONE_TOKEN"
	tokenDefaultLocation = "/etc/kubernetes/oneandone.json"
)

// GetOneandoneToken uses environment variables to locate a 1&1
// token. It will look at a file defined at en environment variable fisrt,
// then to an environment variable
func GetOneandoneToken() (string, error) {
	// try to load from file from env
	if f, ok := os.LookupEnv(tokenFileEnv); ok && f != "" {
		token, err := ReadTokenFromJSONFile(f)
		if err == nil && token != "" {
			return token, nil
		}
		glog.Infof("Could not find a valid configuration file at %s", f)
	}

	// try to load from environment
	if t, ok := os.LookupEnv(tokenEnv); ok {
		token := strings.TrimSpace(t)
		if token != "" {
			return token, nil
		}
		glog.Infof("Could not find a valid token at environment variable %s", tokenEnv)
	}

	//try the default location
	token, err := ReadTokenFromJSONFile(tokenDefaultLocation)
	if err == nil && token != "" {
		return token, nil
	}
	glog.Infof("Could not find a valid configuration file at %s", tokenDefaultLocation)

	return "", fmt.Errorf("No valid 1and1 tokens were found: %s", err)
}

// Config contains 1&1 configuration items
type Config struct {
	Token string `json:"token"`
}

// ReadTokenFromJSONFile reads the 1&1 token from a config file
func ReadTokenFromJSONFile(file string) (string, error) {
	c, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	config := &Config{}
	err = json.Unmarshal(c, config)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(config.Token), nil

}
