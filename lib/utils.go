package lib

import (
	"strconv"

	"github.com/tidwall/buntdb"
)

// GetCounter returns a counter, from 0 to max, stored with the given key (and adds one to it).
func GetCounter(db *buntdb.DB, key string, max int) int {
	var counter int
	_ = db.Update(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err == nil {
			// err doesn't matter here, it'll just return 0 anyway
			counter, _ = strconv.Atoi(val)
		}

		if counter > max || counter < 0 {
			counter = 0
			tx.Set(key, "0", nil)
		} else {
			tx.Set(key, strconv.Itoa(counter+1), nil)
		}

		return nil
	})
	return counter
}
