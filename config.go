package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// SendgridAddressConfig holds the config for a Sendgrid email address
type SendgridAddressConfig struct {
	Name    string
	Address string
}

// NotifyTargetsConfig holds notification targets.
type NotifyTargetsConfig struct {
	SmsTelstra    []string                `yaml:"sms-telstra"`
	EmailSendgrid []SendgridAddressConfig `yaml:"email-sendgrid"`
}

// SmsTelstraConfig holds the configuration for Telsta SMS notifications.
type SmsTelstraConfig struct {
	Key    string
	Secret string
}

// EmailSendgridConfig holds the configuration for Sendgrid email notifications.
type EmailSendgridConfig struct {
	FromName    string `yaml:"from-name"`
	FromAddress string `yaml:"from-address"`
	APIKey      string `yaml:"api-key"`
}

// NotifyConfig holds the configuration for the notifiers.
type NotifyConfig struct {
	DefaultTargets NotifyTargetsConfig `yaml:"default-targets"`
	SmsTelstra     SmsTelstraConfig    `yaml:"sms-telstra"`
	EmailSendgrid  EmailSendgridConfig `yaml:"email-sendgrid"`
}

// WebpageConfig holds the monitor configuration for a web page.
type WebpageConfig struct {
	URL string
}

// Socks5Config holds the monitor configuration for a SOCKS5 proxy.
type Socks5Config struct {
	Host       string
	Port       int
	Username   string
	Password   string
	TestDomain string `yaml:"test-domain"`
}

// Config holds the entire configuration for the service monitor.
type Config struct {
	Notify NotifyConfig

	//TODO(dan): Possibly allow additional notify targets on specific services?
	Services struct {
		Webpage map[string]WebpageConfig
		Socks5  map[string]Socks5Config
	}
}

// LoadConfig loads and returns the Config.
func LoadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
