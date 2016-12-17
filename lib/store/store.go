package store

import (
	"fmt"
	"strconv"
	"time"

	"github.com/LondonTrustMedia/downtime_alert/lib"
	"github.com/tidwall/buntdb"
)

const (
	keyCounter = "counter %s %s"

	keyDowntimeCount            = "ongoing.downtime.count %s %s"
	keyDowntimeLastNotification = "ongoing.last.notification %s %s"
)

var (
	// LatestVersion is the newest version of the data store.
	LatestVersion = 1
)

// DataStore represents a data store.
type DataStore interface {
	Version() int
	Shutdown()

	GetCounter(section, name string, max int)
	MarkDown(section, name string)
	MarkUp(section, name string)
	ShouldAlertDowntime(config lib.OngoingConfig, section, name string)
}

// BuntDBStore is a datastore backed by BuntDB.
type BuntDBStore struct {
	DB           *buntdb.DB
	ShuttingDown chan bool
}

// NewBuntDBStore returns a new BuntDBStore.
func NewBuntDBStore(path string) (*BuntDBStore, error) {
	bunt, err := buntdb.Open(path)
	if err == nil {
		return &BuntDBStore{
			DB:           bunt,
			ShuttingDown: make(chan bool),
		}, nil
	}
	return nil, err
}

// Version returns the latest db version, used for auto-upgrades.
func (store *BuntDBStore) Version(none *interface{}, version *int) error {
	version = &LatestVersion
	return nil
}

// Shutdown shuts down the running db, even remotely.
func (store *BuntDBStore) Shutdown() {
	store.DB.Close()
	store.ShuttingDown <- true
}

// GetCounterArgs are the arguments for GetCounter.
type GetCounterArgs struct {
	Section string
	Name    string
	Max     int
}

// GetCounter returns a counter, from 0 to max, stored with the given key (and adds one to it).
func (store *BuntDBStore) GetCounter(args *GetCounterArgs) int {
	var counter int
	key := fmt.Sprintf(keyCounter, args.Section, args.Name)
	_ = store.DB.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err == nil {
			// err doesn't matter here, it'll just return 0 anyway
			counter, _ = strconv.Atoi(val)
		}

		if counter > args.Max || counter < 0 {
			counter = 0
			tx.Set(key, "0", nil)
		} else {
			tx.Set(key, strconv.Itoa(counter+1), nil)
		}

		return nil
	})
	return counter
}

// MarkArgs are the args for MarkDown/MarkUp.
type MarkArgs struct {
	Section string
	Name    string
}

// MarkDown marks the given service as being down in the datastore.
func (store *BuntDBStore) MarkDown(args *MarkArgs) {
	downtimeCountKey := fmt.Sprintf(keyDowntimeCount, args.Section, args.Name)
	err := store.DB.Update(func(tx *buntdb.Tx) error {
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
func (store *BuntDBStore) MarkUp(args *MarkArgs) {
	downtimeCountKey := fmt.Sprintf(keyDowntimeCount, args.Section, args.Name)
	downtimeLastNotificationKey := fmt.Sprintf(keyDowntimeLastNotification, args.Section, args.Name)
	err := store.DB.Update(func(tx *buntdb.Tx) error {
		tx.Delete(downtimeCountKey)
		tx.Delete(downtimeLastNotificationKey)
		return nil
	})

	if err != nil {
		fmt.Println("Couldn't write update:", err.Error())
	}
}

// ShouldAlertDowntimeArgs are args for ShouldAlertDowntime.
type ShouldAlertDowntimeArgs struct {
	Config           lib.OngoingConfig
	Section          string
	Name             string
	FailsBeforeAlert int
}

// ShouldAlertDowntime returns true if the alerter should send an alert for the given service.
func (store *BuntDBStore) ShouldAlertDowntime(args *ShouldAlertDowntimeArgs) bool {
	var shouldAlert bool

	downtimeCountKey := fmt.Sprintf(keyDowntimeCount, args.Section, args.Name)
	downtimeLastNotificationKey := fmt.Sprintf(keyDowntimeLastNotification, args.Section, args.Name)
	err := store.DB.Update(func(tx *buntdb.Tx) error {
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
			}

			// parse ongoing delay
			ongoingDelay, _ = time.ParseDuration(args.Config.OngoingDelay)
		} else {
			shouldAlert = true
		}

		// see whether to alert, based on options
		if args.FailsBeforeAlert <= downtimeCounts && downtimeCounts <= args.FailsBeforeAlert+args.Config.InitialMaxAlerts {
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
