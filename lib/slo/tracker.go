package slo

import (
	"encoding/json"
	"fmt"
	"time"

	"code.cloudfoundry.org/bytefmt"
)

type DownloadHistoryEntry struct {
	RecordedTime   time.Time `json:"time"`
	Failed         bool
	FailMessage    string `json:"fail-msg"`
	BytesPerSecond uint64 `json:"bytes-per-second"`
}

// DownloadTracker tracks uptime/speed data and SLO objectives.
type DownloadTracker struct {
	History []DownloadHistoryEntry
}

// NewTracker returns a new DownloadTracker.
func NewTracker() *DownloadTracker {
	return &DownloadTracker{}
}

// LoadFromString returns a DownloadTracker instance, from a string representation created by ToString.
func LoadFromString(representation string) (*DownloadTracker, error) {
	var t *DownloadTracker
	err := json.Unmarshal([]byte(representation), &t)
	return t, err
}

// String returns a string representation of DownloadTracker.
func (t *DownloadTracker) String() string {
	trackerString, _ := json.Marshal(t)
	return string(trackerString)
}

// AddDownload adds a successful download to our history.
func (t *DownloadTracker) AddDownload(recordedTime time.Time, bytesPerSecond uint64) {
	t.History = append(t.History, DownloadHistoryEntry{
		RecordedTime:   recordedTime,
		Failed:         false,
		BytesPerSecond: bytesPerSecond,
	})
}

// AddFailure adds a failure entry to our history.
func (t *DownloadTracker) AddFailure(recordedTime time.Time, message string) {
	t.History = append(t.History, DownloadHistoryEntry{
		RecordedTime: recordedTime,
		Failed:       true,
		FailMessage:  message,
	})
}

// CullHistory removes old history entries.
func (t *DownloadTracker) CullHistory(earliestTimeToKeep time.Time) {
	// all good
	if len(t.History) < 1 || t.History[0].RecordedTime.After(earliestTimeToKeep) {
		return
	}

	var newHistory []DownloadHistoryEntry

	for _, info := range t.History {
		if info.RecordedTime.After(earliestTimeToKeep) {
			newHistory = append(newHistory, info)
		}
	}

	t.History = newHistory
}

// TotalTestsPerformed returns how many tests have been performed.
func (t *DownloadTracker) TotalTestsPerformed() int {
	return len(t.History)
}

// SuccessfulTestsPerformed returns how many successful tests have been performed.
// Useful when looking at when to use results from SpeedIsAbove.
func (t *DownloadTracker) SuccessfulTestsPerformed() int {
	var tests int
	for _, info := range t.History {
		if !info.Failed {
			tests++
		}
	}
	return tests
}

// ConsecutiveFailures returns the last consecutive failues and their error messages.
func (t *DownloadTracker) ConsecutiveFailures() (int, []string) {
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
func (t *DownloadTracker) UptimeIsAbove(acceptableUptime float64) bool {
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
func (t *DownloadTracker) SpeedIsAbove(minimumBytesPerSecond uint64, passTarget float64) bool {
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
func (t *DownloadTracker) AverageSpeed() string {
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
