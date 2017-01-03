package main

import (
	"fmt"
	"log"
	"math/rand"

	"time"

	"github.com/LondonTrustMedia/downtime_alert/lib"
	"github.com/LondonTrustMedia/downtime_alert/lib/slo"

	"net"

	"code.cloudfoundry.org/bytefmt"
	"github.com/docopt/docopt-go"
	"github.com/tidwall/buntdb"
)

const (
	keySloTracker = "slo-tracker %s %s"
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

// LoadFromDatastore returns a Tracker instance from the given datastore.
func LoadFromDatastore(db *buntdb.DB, section, name string) (*slo.Tracker, error) {
	sloTrackerKey := fmt.Sprintf(keySloTracker, section, name)
	var tracker *slo.Tracker
	err := db.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(sloTrackerKey)
		if err != nil || len(val) < 1 {
			return err
		}
		tracker, err = slo.LoadFromString(val)
		return err
	})
	return tracker, err
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
			log.Fatal("Could not load config file: ", err.Error())
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

		// seed random numbers (used to uniquify URLs to bypass caches)
		rand.Seed(time.Now().UnixNano())

		// check SOCKS5 proxies
		for name, mconfig := range config.Services.Socks5 {
			// see whether to skip check on this launch
			countWait := lib.GetCounter(db, fmt.Sprintf("socks5-%s-%d-countwait", mconfig.Host, mconfig.Port), mconfig.WaitBetweenAttempts)

			if countWait != 0 {
				log.Println("Skipping SOCKS5 check for", mconfig.Host, "this launch")
				continue
			}

			// get which set of creds to use
			credsToUse := lib.GetCounter(db, fmt.Sprintf("socks5-%s-%d-credentials", mconfig.Host, mconfig.Port), len(mconfig.Credentials)-1)

			// confirm that we have our SLO tracker
			tracker, err := LoadFromDatastore(db, "socks5", name)
			if err != nil {
				tracker = slo.NewTracker()
			}

			// check!
			err = lib.CheckSocks5(tracker, mconfig, credsToUse)
			if err != nil {
				tracker.AddFailure(time.Now(), err.Error())
			}

			// remove old history
			tracker.CullHistory(time.Now().Add(mconfig.TestDownload.SLO.HistoryRetained * -1))

			// check specific failures
			//TODO(dan): Don't alert 3000 times for the same issue, implement failure pattern detection and hiding and all.
			// We'll likely integrate this in as a "ShouldAlert" function into the tracker itself.
			failCount, failMessages := tracker.ConsecutiveFailures()
			if failCount >= mconfig.TestDownload.SLO.MaxFailuresInARow {
				FailAndNotify(config.Notify, name, fmt.Sprintf("Failed %d times in a row:\n%s", failCount, failMessages))
				continue
			}

			if tracker.TotalTestsPerformed() >= 3 && !tracker.UptimeIsAbove(mconfig.TestDownload.SLO.UptimeTarget) {
				FailAndNotify(config.Notify, name, fmt.Sprintf("Uptime is lower than %f", mconfig.TestDownload.SLO.UptimeTarget))
				continue
			}

			if tracker.SuccessfulTestsPerformed() >= 3 && !tracker.SpeedIsAbove(mconfig.TestDownload.SLO.MinBytesPerSecond, mconfig.TestDownload.SLO.SpeedTarget) {
				FailAndNotify(config.Notify, name, fmt.Sprintf("Proxy is very slow (target of %s/s for %d%% of connections not met)", bytefmt.ByteSize(mconfig.TestDownload.SLO.MinBytesPerSecond), int(mconfig.TestDownload.SLO.SpeedTarget*100)))
				continue
			}

			// save tracker
			sloTrackerKey := fmt.Sprintf(keySloTracker, "socks5", name)
			db.Update(func(tx *buntdb.Tx) error {
				tx.Set(sloTrackerKey, tracker.String(), nil)
				return nil
			})
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
