package main

import (
	"fmt"
	"log"

	"time"

	"github.com/LondonTrustMedia/downtime_alert/lib"

	"net"

	"github.com/docopt/docopt-go"
	"github.com/tidwall/buntdb"
)

// FailAndNotify notifies about the failure using whatever methods have been selected and errors out.
func FailAndNotify(nconfig lib.NotifyConfig, serviceName string, errorMessage string) {
	message := fmt.Sprintf("== %s is down ==\n%s", serviceName, errorMessage)
	log.Println(message)

	// send Telstra SMS to the given phone numbers.
	for _, phoneNumber := range nconfig.DefaultTargets.SmsTelstra {
		log.Println("Sending SMS notification of failure to", phoneNumber)
		lib.SendSMSTelstra(nconfig.SmsTelstra.Key, nconfig.SmsTelstra.Secret, phoneNumber, message)
	}

	// send Sendgrid emails to the given targets.
	if len(nconfig.DefaultTargets.EmailSendgrid) > 0 {
		log.Println("Sending email notification of failure to", nconfig.DefaultTargets.EmailSendgrid)
		lib.SendEmailSendgrid(nconfig.EmailSendgrid.APIKey, nconfig.EmailSendgrid.FromName, nconfig.EmailSendgrid.FromAddress, nconfig.DefaultTargets.EmailSendgrid, message)
	}
}

func main() {
	usage := `downtimealert.
downtimealert connects to and monitors services, and reports outages.

Usage:
	downtimealert try [--config=<filename>] [--onecopy]
	downtimealert -h | --help
	downtimealert --version

Options:
	--config=<filename>    Use the given config file [default: config.yaml].
	--onecopy              Ensure that only one copy is running at a time.

	-h --help    Show this screen.
	--version    Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, fmt.Sprintf("downtimealert v%s", lib.SemVer), false)

	if arguments["try"].(bool) {
		log.Println("Trying services")

		// load config
		config, err := lib.LoadConfig(arguments["--config"].(string))
		if err != nil {
			log.Fatal("Could not load config file:", err.Error())
		}

		// ensure only one copy of this alerter exists
		if arguments["--onecopy"].(bool) {
			onecopy, err := net.Listen("tcp", config.Onecopy)
			if err != nil {
				// a downtimealert is already running, so exit
				log.Println("Downtime alerter already running, exiting")
				return
			}
			log.Println("Opened onecopy listener at", config.Onecopy)
			defer onecopy.Close()
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

			// get which set of creds to use
			credsToUse := lib.GetCounter(db, fmt.Sprintf("socks5-%s-%d-credentials", mconfig.Host, mconfig.Port), len(mconfig.Credentials)-1)

			err = lib.CheckSocks5(mconfig, credsToUse)
			if err != nil {
				failure = true
			}

			if failure {
				lib.MarkDown(db, "socks5", name)

				// if we should alert the customer, go yell at them
				if lib.ShouldAlertDowntime(db, config.Ongoing, "socks5", name, 3) {
					FailAndNotify(config.Notify, name, fmt.Sprintf("Host: %s\nError: %s", mconfig.Host, err.Error()))
				}
			} else {
				lib.MarkUp(db, "socks5", name)
			}
		}

		// check web pages
		for name, mconfig := range config.Services.Webpage {
			// require two failures in a row to report it, to prevent notification on momentary net glitches
			var failure bool

			err = lib.CheckWebpage(mconfig)
			if err != nil {
				// wait for momentary net glitches to pass
				time.Sleep(config.RecheckDelayDuration)
				err = lib.CheckWebpage(mconfig)
				if err != nil {
					failure = true
				}
			}

			if failure {
				lib.MarkDown(db, "webpage", name)

				// if we should alert the customer, go yell at them
				if lib.ShouldAlertDowntime(db, config.Ongoing, "webpage", name, 2) {
					FailAndNotify(config.Notify, name, fmt.Sprintf("URL: %s\nStatus: %s", mconfig.URL, err.Error()))
				}
			} else {
				lib.MarkUp(db, "webpage", name)
			}
		}
	}
}
