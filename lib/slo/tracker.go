package slo

import "time"
import "encoding/json"

type historyEntry struct {
	RecordedTime   time.Time `json:"time"`
	Failed         bool
	FailMessage    string `json:"fail-msg"`
	BytesPerSecond uint64 `json:"bytes-per-second"`
}

// Tracker tracks uptime/speed data and SLO objectives.
type Tracker struct {
	history []historyEntry
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
	t.history = append(t.history, historyEntry{
		RecordedTime:   recordedTime,
		Failed:         false,
		BytesPerSecond: bytesPerSecond,
	})
}

// AddFailure adds a failure entry to our history.
func (t *Tracker) AddFailure(recordedTime time.Time, message string) {
	t.history = append(t.history, historyEntry{
		RecordedTime: recordedTime,
		Failed:       true,
		FailMessage:  message,
	})
}

// CullHistory removes old history entries.
func (t *Tracker) CullHistory(earliestTimeToKeep time.Time) {
	// all good
	if len(t.history) < 1 || t.history[0].RecordedTime.After(earliestTimeToKeep) {
		return
	}

	var newHistory []historyEntry

	for _, info := range t.history {
		if info.RecordedTime.After(earliestTimeToKeep) {
			newHistory = append(newHistory, info)
		}
	}

	t.history = newHistory
}

// TotalTestsPerformed returns how many tests have been performed.
func (t *Tracker) TotalTestsPerformed() int {
	return len(t.history)
}

// SuccessfulTestsPerformed returns how many successful tests have been performed.
// Useful when looking at when to use results from SpeedIsAbove.
func (t *Tracker) SuccessfulTestsPerformed() int {
	var tests int
	for _, info := range t.history {
		if !info.Failed {
			tests++
		}
	}
	return tests
}

// ConsecutiveFailures returns the last consecutive failues and their error messages.
func (t *Tracker) ConsecutiveFailures() (int, []string) {
	if len(t.history) < 1 || !t.history[len(t.history)-1].Failed {
		return 0, []string{}
	}

	// not efficient, but it works and is simple to implement
	var failErrorMessages []string
	for _, info := range t.history {
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
	if len(t.history) < 1 {
		return true
	}

	// generate uptime
	var failedTests int
	var overallTests int

	for _, info := range t.history {
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
	if len(t.history) < 1 {
		return true
	}

	// generate uptime
	var failedTests int
	var overallTests int

	for _, info := range t.history {
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
