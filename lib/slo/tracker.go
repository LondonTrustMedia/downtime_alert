package slo

import (
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/bytefmt"
)

type HistoryEntry struct {
	RecordedTime   time.Time `json:"time"`
	Failed         bool
	FailMessage    string `json:"fail-msg"`
	BytesPerSecond uint64 `json:"bytes-per-second"`
}

// Tracker tracks uptime/speed data and SLO objectives.
type Tracker struct {
	History []HistoryEntry
}

// NewTracker returns a new Tracker.
func NewTracker() *Tracker {
	return &Tracker{}
}

// LoadFromString returns a Tracker instance, from a string representation created by ToString.
func LoadFromString(representation string) (*Tracker, error) {
	var t *Tracker
	err := json.Unmarshal([]byte(representation), &t)
	return t, err
}

// String returns a string representation of Tracker.
func (t *Tracker) String() string {
	trackerString, _ := json.Marshal(t)
	return string(trackerString)
}

// AddDownload adds a successful download to our history.
func (t *Tracker) AddDownload(recordedTime time.Time, bytesPerSecond uint64) {
	t.History = append(t.History, HistoryEntry{
		RecordedTime:   recordedTime,
		Failed:         false,
		BytesPerSecond: bytesPerSecond,
	})
}

// AddFailure adds a failure entry to our history.
func (t *Tracker) AddFailure(recordedTime time.Time, message string) {
	t.History = append(t.History, HistoryEntry{
		RecordedTime: recordedTime,
		Failed:       true,
		FailMessage:  message,
	})
}

// CullHistory removes old history entries.
func (t *Tracker) CullHistory(earliestTimeToKeep time.Time) {
	// all good
	if len(t.History) < 1 || t.History[0].RecordedTime.After(earliestTimeToKeep) {
		return
	}

	var newHistory []HistoryEntry

	for _, info := range t.History {
		if info.RecordedTime.After(earliestTimeToKeep) {
			newHistory = append(newHistory, info)
		}
	}

	t.History = newHistory
}

// TotalTestsPerformed returns how many tests have been performed.
func (t *Tracker) TotalTestsPerformed() int {
	return len(t.History)
}

// SuccessfulTestsPerformed returns how many successful tests have been performed.
// Useful when looking at when to use results from SpeedIsAbove.
func (t *Tracker) SuccessfulTestsPerformed() int {
	var tests int
	for _, info := range t.History {
		if !info.Failed {
			tests++
		}
	}
	return tests
}

// ConsecutiveFailures returns the last consecutive failues and their error messages.
func (t *Tracker) ConsecutiveFailures() (int, []string) {
	if len(t.History) < 1 || !t.History[len(t.History)-1].Failed {
		return 0, []string{}
	}

	// not efficient, but it works and is simple to implement
	var failErrorMessages []string
	for _, info := range t.History {
		if !info.Failed {
			failErrorMessages = []string{}
			continue
		}

		failErrorMessages = append(failErrorMessages, info.FailMessage)
	}

	return len(failErrorMessages), failErrorMessages
}

// UptimeIsAbove says whether the current uptime is above the given percentage.
func (t *Tracker) UptimeIsAbove(acceptableUptime float64) bool {
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

// SpeedIsAbove says whether the current uptime is above the given percentage.
func (t *Tracker) SpeedIsAbove(minimumBytesPerSecond uint64, passTarget float64) bool {
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
		if info.BytesPerSecond < minimumBytesPerSecond {
			failedTests++
		}
	}

	passedTests := float64(1) - (float64(failedTests) / float64(overallTests))

	if passedTests > passTarget {
		return true
	}
	return false
}

// AverageSpeed returns the average speed of all the tests we're keeping track of.
func (t *Tracker) AverageSpeed() string {
	var overallSpeed uint64
	var overallTests int

	for _, info := range t.History {
		if info.Failed {
			continue
		}
		overallTests++
		overallSpeed += info.BytesPerSecond
	}

	averageSpeedInBytes := overallSpeed / uint64(overallTests)

	return fmt.Sprintf("%s/s", bytefmt.ByteSize(averageSpeedInBytes))
}
