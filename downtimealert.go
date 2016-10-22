package main

import (
	"fmt"
	"log"

	"github.com/LondonTrustMedia/downtime_alert/da"
	"github.com/docopt/docopt-go"
	"github.com/tidwall/buntdb"
)

// FailAndNotify notifies about the failure using whatever methods have been selected and errors out.
func FailAndNotify(nconfig da.NotifyConfig, serviceName string, errorMessage string) {
	message := fmt.Sprintf("== %s is down ==\n%s", serviceName, errorMessage)
	log.Println(message)

	// send Telstra SMS to the given phone numbers.
	for _, phoneNumber := range nconfig.DefaultTargets.SmsTelstra {
		log.Println("Sending SMS notification of failure to", phoneNumber)
		da.SendSMSTelstra(nconfig.SmsTelstra.Key, nconfig.SmsTelstra.Secret, phoneNumber, message)
	}

	// send Sendgrid emails to the given targets.
	if len(nconfig.DefaultTargets.EmailSendgrid) > 0 {
		log.Println("Sending email notification of failure to", nconfig.DefaultTargets.EmailSendgrid)
		da.SendEmailSendgrid(nconfig.EmailSendgrid.APIKey, nconfig.EmailSendgrid.FromName, nconfig.EmailSendgrid.FromAddress, nconfig.DefaultTargets.EmailSendgrid, message)
	}
}

func main() {
	version := "downtimealert 0.1.0"
	usage := `downtimealert.
downtimealert connects to and monitors services, and reports outages.

Usage:
	downtimealert try [--config=<filename>]
	downtimealert -h | --help
	downtimealert --version

Options:
	--config=<filename>    Use the given config file [default: config.yaml].

	-h --help    Show this screen.
	--version    Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, version, false)

	if arguments["try"].(bool) {
		log.Println("Trying services")

		// load config
		config, err := da.LoadConfig(arguments["--config"].(string))
		if err != nil {
			log.Fatal("Could not load config file: %s", err.Error())
		}

		// load datastore
		db, err := buntdb.Open(config.Datastore)
		if err != nil {
			FailAndNotify(config.Notify, "Datastore", fmt.Sprintf("Couldn't open bunt datastore: %s", err.Error()))
		}
		defer db.Close()

		// check SOCKS5 proxies
		for name, mconfig := range config.Services.Socks5 {
			// require two failures in a row to report it, to prevent notification on momentary net glitches
			var failure bool
			err = da.CheckSocks5(mconfig)
			if err != nil {
				err = da.CheckSocks5(mconfig)
				if err != nil {
					failure = true
				}
			}

			if failure {
				da.MarkDown(db, "socks5", name)

				// if we should alert the customer, go yell at them
				if da.ShouldAlertDowntime(db, config.Ongoing, "socks5", name) {
					FailAndNotify(config.Notify, name, fmt.Sprintf("Host: %s\nError: %s", mconfig.Host, err.Error()))
				}
			} else {
				da.MarkUp(db, "socks5", name)
			}
		}

		// check web pages
		for name, mconfig := range config.Services.Webpage {
			// require two failures in a row to report it, to prevent notification on momentary net glitches
			var failure bool

			err = da.CheckWebpage(mconfig)
			if err != nil {
				err = da.CheckWebpage(mconfig)
				if err != nil {
					failure = true
				}
			}

			if failure {
				da.MarkDown(db, "webpage", name)

				// if we should alert the customer, go yell at them
				if da.ShouldAlertDowntime(db, config.Ongoing, "webpage", name) {
					FailAndNotify(config.Notify, name, fmt.Sprintf("URL: %s\nStatus: %s", mconfig.URL, err.Error()))
				}
			} else {
				da.MarkUp(db, "webpage", name)
			}
		}
	}
}
