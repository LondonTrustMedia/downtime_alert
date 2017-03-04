package lib

import (
	"fmt"
	"io/ioutil"

	"time"

	"code.cloudfoundry.org/bytefmt"
	"gopkg.in/yaml.v2"
)

// OngoingConfig holds the configuration used for ongoing issues.
type OngoingConfig struct {
	InitialMaxAlerts int    `yaml:"initial-max-alerts"`
	OngoingDelay     string `yaml:"ongoing-delay"`
}

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
	URL       string
	UserAgent string `yaml:"user-agent"`
	Matches   []string
}

// UserPassCredentialConfig holds credentials for typical username+password services.
type UserPassCredentialConfig struct {
	Username string
	Password string
}

// TestDownloadConfig is the info for a test download.
type TestDownloadConfig struct {
	URL               string
	MaxSizeToDLString string `yaml:"max-size-to-dl"`
	MaxBytesToDL      uint64
	SLO               struct {
		HistoryRetainedString   string `yaml:"history-retained"`
		HistoryRetained         time.Duration
		MaxFailuresInARow       int     `yaml:"max-failures-in-a-row"`
		UptimeTarget            float64 `yaml:"uptime-target"`
		MinSpeedPerSecondString string  `yaml:"min-speed-per-second"`
		MinBytesPerSecond       uint64
		SpeedTarget             float64 `yaml:"speed-target"`
	}
}

// Socks5Config holds the monitor configuration for a SOCKS5 proxy.
type Socks5Config struct {
	Host                string
	Port                int
	WaitBetweenAttempts int `yaml:"wait-between-attempts"`
	Credentials         []UserPassCredentialConfig
	TestDownload        TestDownloadConfig `yaml:"test-download"`
}

// Config holds the entire configuration for the service monitor.
type Config struct {
	Datastore string

	Onecopy string

	RecheckDelay string `yaml:"recheck-delay"`

	RecheckDelayDuration time.Duration

	Ongoing OngoingConfig

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

	// get RecheckDelayDuration
	config.RecheckDelayDuration, err = time.ParseDuration(config.RecheckDelay)
	if err != nil {
		return &config, fmt.Errorf("Could not parse RecheckDelay: %s", err.Error())
	}

	// calculate TestDownloadConfig stuff
	for name, info := range config.Services.Socks5 {
		info.TestDownload.SLO.HistoryRetained, err = time.ParseDuration(info.TestDownload.SLO.HistoryRetainedString)
		if err != nil {
			return &config, fmt.Errorf("Could not parse history-retained in SOCKS5 %s: %s", name, err.Error())
		}

		if info.TestDownload.MaxSizeToDLString != "" {
			info.TestDownload.MaxBytesToDL, err = bytefmt.ToBytes(info.TestDownload.MaxSizeToDLString)
			if err != nil {
				return &config, fmt.Errorf("Could not parse max-size-to-dl in SOCKS5 %s: %s", name, err.Error())
			}
		}

		if info.TestDownload.SLO.MinSpeedPerSecondString != "" {
			info.TestDownload.SLO.MinBytesPerSecond, err = bytefmt.ToBytes(info.TestDownload.SLO.MinSpeedPerSecondString)
			if err != nil {
				return &config, fmt.Errorf("Could not parse min-speed-per-second in SOCKS5 %s: [%s] %s", name, info.TestDownload.SLO.MinSpeedPerSecondString, err.Error())
			}
		}

		// save new info
		config.Services.Socks5[name] = info
	}

	return &config, nil
}
