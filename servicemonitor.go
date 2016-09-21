package main

import (
	"fmt"
	"log"

	"github.com/docopt/docopt-go"
)

// FailAndNotify notifies about the failure using whatever methods have been selected and errors out.
func FailAndNotify(nconfig NotifyConfig, serviceName string, errorMessage string) {
	message := fmt.Sprintf("== %s is down ==\n%s", serviceName, errorMessage)
	log.Printf(message)

	// send Telstra SMS to the given phone numbers.
	for _, phoneNumber := range nconfig.DefaultTargets.SmsTelstra {
		log.Println("Sending SMS notification of failure to", phoneNumber)
		SendSMS(nconfig.SmsTelstra.Key, nconfig.SmsTelstra.Secret, phoneNumber, message)
	}

	//TODO(dan): Alert via email service as well.
}

func main() {
	version := "servicemonitor 0.1.0"
	usage := `servicemonitor.
servicemonitor connects to and monitors services, and reports outages.

Usage:
	servicemonitor try [--config=<filename>]
	servicemonitor -h | --help
	servicemonitor --version

Options:
	--config=<filename>    Use the given config file [default: config.yaml].

	-h --help    Show this screen.
	--version    Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, version, false)

	if arguments["try"].(bool) {
		log.Println("Trying services")

		// load config
		config, err := LoadConfig(arguments["--config"].(string))
		if err != nil {
			log.Fatal("Could not load config file: %s", err.Error())
		}

		// check SOCKS5 proxies
		for name, mconfig := range config.Services.Socks5 {
			// require two failures in a row to report it, to prevent notification on momentary net glitches
			err = CheckSocks5(mconfig)
			if err != nil {
				err = CheckSocks5(mconfig)
				if err != nil {
					FailAndNotify(config.Notify, name, fmt.Sprintf("Host: %s\nError: %s", mconfig.Host, err.Error()))
				}
			}
		}

		// check web pages
		for name, mconfig := range config.Services.Webpage {
			// require two failures in a row to report it, to prevent notification on momentary net glitches
			err = CheckWebpage(mconfig)
			if err != nil {
				err = CheckWebpage(mconfig)
				if err != nil {
					FailAndNotify(config.Notify, name, fmt.Sprintf("URL: %s\nStatus: %s", mconfig.URL, err.Error()))
				}
			}
		}
	}
}
