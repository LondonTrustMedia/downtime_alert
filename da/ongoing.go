package da

import (
	"github.com/tidwall/buntdb"
)

const (
	keyDowntimeCount            = "ongoing.downtime.count %s %s"
	keyDowntimeLastNotification = "ongoing.last.notification %s %s"
)

// MarkDown marks the given service as being down in the datastore.
func MarkDown(db *buntdb.DB, section string, name string) {

}

// MarkUp marks the given service as being up in the datastore.
func MarkUp(db *buntdb.DB, section string, name string) {

}

// ShouldAlertDowntime returns true if the alerter should send an alert for the given service.
func ShouldAlertDowntime(db *buntdb.DB, config OngoingConfig, section string, name string) bool {
	return true
}
