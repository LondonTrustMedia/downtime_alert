package da

import (
	"fmt"
	"strconv"

	"time"

	"github.com/tidwall/buntdb"
)

const (
	keyDowntimeCount            = "ongoing.downtime.count %s %s"
	keyDowntimeLastNotification = "ongoing.last.notification %s %s"
)

// MarkDown marks the given service as being down in the datastore.
func MarkDown(db *buntdb.DB, section string, name string) {
	downtimeCountKey := fmt.Sprintf(keyDowntimeCount, section, name)
	err := db.Update(func(tx *buntdb.Tx) error {
		var lastCount int
		val, err := tx.Get(downtimeCountKey)
		if err == nil {
			// err doesn't matter here, it'll just return 0 anyway
			lastCount, _ = strconv.Atoi(val)
		}

		tx.Set(downtimeCountKey, strconv.Itoa(lastCount+1), nil)

		return nil
	})

	if err != nil {
		fmt.Println("Couldn't write update:", err.Error())
	}
}

// MarkUp marks the given service as being up in the datastore.
func MarkUp(db *buntdb.DB, section string, name string) {
	downtimeCountKey := fmt.Sprintf(keyDowntimeCount, section, name)
	downtimeLastNotificationKey := fmt.Sprintf(keyDowntimeLastNotification, section, name)
	err := db.Update(func(tx *buntdb.Tx) error {
		tx.Delete(downtimeCountKey)
		tx.Delete(downtimeLastNotificationKey)
		return nil
	})

	if err != nil {
		fmt.Println("Couldn't write update:", err.Error())
	}
}

// ShouldAlertDowntime returns true if the alerter should send an alert for the given service.
func ShouldAlertDowntime(db *buntdb.DB, config OngoingConfig, section string, name string) bool {
	var shouldAlert bool

	downtimeCountKey := fmt.Sprintf(keyDowntimeCount, section, name)
	downtimeLastNotificationKey := fmt.Sprintf(keyDowntimeLastNotification, section, name)
	err := db.Update(func(tx *buntdb.Tx) error {
		var downtimeCounts int
		val, err := tx.Get(downtimeCountKey)
		if err == nil {
			// err doesn't matter here, it'll just return 0 anyway
			downtimeCounts, _ = strconv.Atoi(val)
		}

		var lastAlerted time.Time
		var ongoingDelay time.Duration
		val, err = tx.Get(downtimeLastNotificationKey)
		if err == nil {
			// parse last alerted time
			i, err := strconv.ParseInt(val, 10, 64)
			if err == nil {
				lastAlerted = time.Unix(i, 0)
			} else {
				shouldAlert = true
			}

			// parse ongoing delay
			ongoingDelay, _ = time.ParseDuration(config.OngoingDelay)
		} else {
			shouldAlert = true
		}

		// see whether to alert, based on options
		if downtimeCounts <= config.InitialMaxAlerts {
			shouldAlert = true
		} else if !shouldAlert && time.Now().After(lastAlerted.Add(ongoingDelay)) {
			shouldAlert = true
		}

		// update last notification time key
		if shouldAlert {
			tx.Set(downtimeLastNotificationKey, strconv.FormatInt(time.Now().Unix(), 10), nil)
		}

		return nil
	})

	if err != nil {
		fmt.Println("Couldn't write update:", err.Error())
	}
	return shouldAlert
}
