package lib

import (
	"fmt"
	"log"
	"time"

	"github.com/LondonTrustMedia/downtime_alert/lib/slo"
	ping "github.com/sparrc/go-ping"
)

// CheckPing checks the given host and tracks results in the tracker.
func CheckPing(tracker *slo.PingTracker, config PingConfig) error {
	log.Println("Checking PING on", config.Host)

	// test ping connectivity
	pinger, err := ping.NewPinger(config.Host)
	if err != nil {
		return err
	}

	pinger.Count = config.PingsPerRun

	pinger.Run()                 // blocks until finished
	stats := pinger.Statistics() // get send/receive/rtt stats

	for _, rtt := range stats.Rtts {
		tracker.AddPing(time.Now(), rtt)
	}
	for i := 1; i <= stats.PacketsSent-stats.PacketsRecv; i++ {
		tracker.AddFailure(time.Now())
	}

	log.Println("Pinged", config.Host, "-", stats.PacketsSent, "sent,", stats.PacketsRecv, "received, lost", fmt.Sprintf("%.2f%%", stats.PacketLoss*100))

	// no errors!
	return nil
}
