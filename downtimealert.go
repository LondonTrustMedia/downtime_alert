package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/LondonTrustMedia/downtime_alert/lib"
	"github.com/LondonTrustMedia/downtime_alert/lib/store"

	"github.com/docopt/docopt-go"
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
	downtimealert try [--config=<filename>] [--serve-data-store]
	downtimealert -h | --help
	downtimealert --version

Options:
	--config=<filename>    Use the given config file [default: config.yaml].
	--serve-data-store     If there isn't an existing one open, stays open to serve
	                       the data store to other downtimealert instances.

	-h --help    Show this screen.
	--version    Show version.`

	arguments, _ := docopt.Parse(usage, nil, true, fmt.Sprintf("downtimealert v%s", lib.SemVer), false)

	if arguments["try"].(bool) {

		// load config
		config, err := lib.LoadConfig(arguments["--config"].(string))
		if err != nil {
			log.Fatal("Could not load config file:", err.Error())
		}

		// start datastore
		log.Println("Starting data store")

		var openedDataStore bool
		var servedStore *store.BuntDBStore

		l, e := net.Listen("tcp", config.Datastore.URL)
		if e == nil {
			// open datastore
			servedStore, err = store.NewBuntDBStore(config.Datastore.Path)
			if err != nil {
				FailAndNotify(config.Notify, "Datastore", fmt.Sprintf("Couldn't open datastore: %s", err.Error()))
			}

			// serve datastore
			server := rpc.NewServer()
			server.RegisterName("DB", servedStore)

			go server.Accept(l)
			openedDataStore = true
		}

		// open rpc
		// if version is old, kill the old datastore and die
		conn, err := net.Dial("tcp", config.Datastore.URL)
		if err != nil {
			FailAndNotify(config.Notify, "Datastore", fmt.Sprintf("Couldn't connect to datastore: %s", err.Error()))
		}

		db := &store.RPCStore{Client: rpc.NewClient(conn)}

		fmt.Println(db.Version())

		return

		log.Println("Trying services", openedDataStore)

		// // check SOCKS5 proxies
		// for name, mconfig := range config.Services.Socks5 {
		// 	// require two failures in a row to report it, to prevent notification on momentary net glitches
		// 	var failure bool

		// 	// get which set of creds to use
		// 	credsToUse := lib.GetCounter(db, fmt.Sprintf("socks5-%s-%d-credentials", mconfig.Host, mconfig.Port), len(mconfig.Credentials)-1)

		// 	err = lib.CheckSocks5(mconfig, credsToUse)
		// 	if err != nil {
		// 		failure = true
		// 	}

		// 	if failure {
		// 		lib.MarkDown(db, "socks5", name)

		// 		// if we should alert the customer, go yell at them
		// 		if lib.ShouldAlertDowntime(db, config.Ongoing, "socks5", name, 3) {
		// 			FailAndNotify(config.Notify, name, fmt.Sprintf("Host: %s\nError: %s", mconfig.Host, err.Error()))
		// 		}
		// 	} else {
		// 		lib.MarkUp(db, "socks5", name)
		// 	}
		// }

		// // check web pages
		// for name, mconfig := range config.Services.Webpage {
		// 	// require two failures in a row to report it, to prevent notification on momentary net glitches
		// 	var failure bool

		// 	err = lib.CheckWebpage(mconfig)
		// 	if err != nil {
		// 		// wait for momentary net glitches to pass
		// 		time.Sleep(config.RecheckDelayDuration)
		// 		err = lib.CheckWebpage(mconfig)
		// 		if err != nil {
		// 			failure = true
		// 		}
		// 	}

		// 	if failure {
		// 		lib.MarkDown(db, "webpage", name)

		// 		// if we should alert the customer, go yell at them
		// 		if lib.ShouldAlertDowntime(db, config.Ongoing, "webpage", name, 2) {
		// 			FailAndNotify(config.Notify, name, fmt.Sprintf("URL: %s\nStatus: %s", mconfig.URL, err.Error()))
		// 		}
		// 	} else {
		// 		lib.MarkUp(db, "webpage", name)
		// 	}
		// }
	}
}
