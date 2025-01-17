package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/creasty/defaults"
)

type CloudAuth struct {
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	ProjectName       string `yaml:"project_name"`
	ProjectDomainName string `yaml:"project_domain_name"`
	UserDomainName    string `yaml:"user_domain_name"`
	AuthURL           string `yaml:"auth_url"`
	Verify            bool   `yaml:"verify,omitempty" default:"true"`
	CACert            string `yaml:"cacert,omitempty"`
}

type Cloud struct {
	Region             string    `yaml:"region_name"`
	IdentityAPIVersion string    `yaml:"identity_api_version"`
	IdentityInterface  string    `yaml:"identity_interface"`
	Auth               CloudAuth `yaml:"auth"`
}

type CloudConfig struct {
	Clouds map[string]Cloud `yaml:"clouds"`
}

func (config *Cloud) GetTLSConfig() (*tls.Config, error) {
	var tlsConfig tls.Config

	if config.Auth.CACert == "" && config.Auth.Verify {
		return nil, nil
	}

	if !config.Auth.Verify {
		tlsConfig.InsecureSkipVerify = true
	}

	if config.Auth.CACert != "" {
		caCertPool := x509.NewCertPool()
		if ok := caCertPool.AppendCertsFromPEM([]byte(config.Auth.CACert)); !ok {
			return nil, fmt.Errorf("unable to load cacert")
		}
		tlsConfig.RootCAs = caCertPool
	}

	return &tlsConfig, nil
}

func (config *Cloud) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type cloud Cloud
	var tmpConfig cloud

	// Set the default values for config structure.
	if err := defaults.Set(&tmpConfig); err != nil {
		return err
	}

	if err := unmarshal(&tmpConfig); err != nil {
		return err
	}

	*config = Cloud(tmpConfig)
	return nil
}

func (config *CloudConfig) GetByName(name string) (*Cloud, error) {
	cloud, ok := config.Clouds[name]
	if !ok {
		return nil, fmt.Errorf("cloud %s not found", name)
	}

	return &cloud, nil
}

func NewCloudConfigFromByteArray(data []byte) (*CloudConfig, error) {
	var config CloudConfig

	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, err
}

func NewCloudConfigFromFile(file string) (*CloudConfig, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return NewCloudConfigFromByteArray(data)
}
