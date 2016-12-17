package store

import (
	"net/rpc"

	"github.com/LondonTrustMedia/downtime_alert/lib"
)

// RPCStore is used to convey stuff to the RPC side.
type RPCStore struct {
	Client *rpc.Client
}

func (db *RPCStore) Version() int {
	var reply int
	db.Client.Call("DB.Version", nil, &reply)
	return reply
}

func (db *RPCStore) Shutdown() {
	db.Client.Call("DB.Shutdown", nil, nil)
}

func (db *RPCStore) GetCounter(section, name string, max int) {
	db.Client.Call("DB.GetCounter", &GetCounterArgs{
		Section: section,
		Name:    name,
		Max:     max,
	}, nil)
}

func (db *RPCStore) MarkDown(section, name string) {
	db.Client.Call("DB.MarkDown", &MarkArgs{
		Section: section,
		Name:    name,
	}, nil)
}

func (db *RPCStore) MarkUp(section, name string) {
	db.Client.Call("DB.MarkUp", &MarkArgs{
		Section: section,
		Name:    name,
	}, nil)
}

func (db *RPCStore) ShouldAlertDowntime(config lib.OngoingConfig, section, name string, failsBeforeAlert int) bool {
	var shouldAlert bool
	db.Client.Call("DB.ShouldAlertDowntime", &ShouldAlertDowntimeArgs{
		Config:           config,
		Section:          section,
		Name:             name,
		FailsBeforeAlert: failsBeforeAlert,
	}, &shouldAlert)
	return shouldAlert
}
