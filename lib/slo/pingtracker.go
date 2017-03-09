package slo

import (
	"encoding/json"
	"time"
)

// PingHistoryEntry represents a single ping, whether it succeeded or failed, and the RTT.
type PingHistoryEntry struct {
	RecordedTime time.Time `json:"time"`
	Failed       bool
	RTT          time.Duration `json:"rtt"`
}

// PingTracker tracks uptime/speed data and SLO objectives.
type PingTracker struct {
	History []PingHistoryEntry
}

// NewPingTracker returns a new PingTracker.
func NewPingTracker() *PingTracker {
	return &PingTracker{}
}

// LoadPingTrackerFromString returns a PingTracker instance, from a string representation created by ToString.
func LoadPingTrackerFromString(representation string) (*PingTracker, error) {
	var t *PingTracker
	err := json.Unmarshal([]byte(representation), &t)
	return t, err
}

// String returns a string representation of PingTracker.
func (t *PingTracker) String() string {
	trackerString, _ := json.Marshal(t)
	return string(trackerString)
}

// AddPing adds a successful ping to our history.
func (t *PingTracker) AddPing(recordedTime time.Time, rtt time.Duration) {
	t.History = append(t.History, PingHistoryEntry{
		RecordedTime: recordedTime,
		Failed:       false,
		RTT:          rtt,
	})
}

// AddFailure adds a failure entry to our history.
func (t *PingTracker) AddFailure(recordedTime time.Time) {
	t.History = append(t.History, PingHistoryEntry{
		RecordedTime: recordedTime,
		Failed:       true,
	})
}

// CullHistory removes old history entries.
func (t *PingTracker) CullHistory(earliestTimeToKeep time.Time) {
	// all good
	if len(t.History) < 1 || t.History[0].RecordedTime.After(earliestTimeToKeep) {
		return
	}

	var newHistory []PingHistoryEntry

	for _, info := range t.History {
		if info.RecordedTime.After(earliestTimeToKeep) {
			newHistory = append(newHistory, info)
		}
	}

	t.History = newHistory
}

// TotalTestsPerformed returns how many tests have been performed.
func (t *PingTracker) TotalTestsPerformed() int {
	return len(t.History)
}

// SuccessfulTestsPerformed returns how many successful tests have been performed.
// Useful when looking at when to use results from SpeedIsAbove.
func (t *PingTracker) SuccessfulTestsPerformed() int {
	var tests int
	for _, info := range t.History {
		if !info.Failed {
			tests++
		}
	}
	return tests
}

// ConsecutiveFailures returns the last consecutive failues and their error messages.
func (t *PingTracker) ConsecutiveFailures() int {
	if len(t.History) < 1 || !t.History[len(t.History)-1].Failed {
		return 0
	}

	var fails int
	for _, info := range t.History {
		if !info.Failed {
			continue
		}

		fails++
	}

	return fails
}

// UptimeIsAbove says whether the current uptime is above the given percentage.
func (t *PingTracker) UptimeIsAbove(acceptableUptime float64) bool {
	if len(t.History) < 1 {
		return true
	}

	// generate uptime
	var failedTests int
	var overallTests int

	for _, info := range t.History {
		overallTests++
		if info.Failed {
			failedTests++
		}
	}

	passedTests := float64(1) - (float64(failedTests) / float64(overallTests))

	if passedTests > acceptableUptime {
		return true
	}
	return false
}

// AvgRTTIsBelow says whether the average RTT is above the given duration.
func (t *PingTracker) AvgRTTIsBelow(maximumRTT time.Duration, passTarget float64) bool {
	if len(t.History) < 1 {
		return true
	}

	// generate uptime
	var failedTests int
	var overallTests int

	for _, info := range t.History {
		if info.Failed {
			continue
		}
		overallTests++
		if info.RTT > maximumRTT {
			failedTests++
		}
	}

	passedTests := float64(1) - (float64(failedTests) / float64(overallTests))

	if passedTests > passTarget {
		return true
	}
	return false
}

// AverageRTT returns the average RTT of all the tests we're keeping track of.
func (t *PingTracker) AverageRTT() time.Duration {
	var overallRTT time.Duration
	var overallTests int

	for _, info := range t.History {
		if info.Failed {
			continue
		}
		overallTests++
		overallRTT += info.RTT
	}

	// this might be the recommended way but it feels very hacky
	return overallRTT / time.Duration(overallTests)
}
