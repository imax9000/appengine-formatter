package appengine

import (
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestErrorNotLost(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("error", errors.New("wild walrus")))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["error"] != "wild walrus" {
		t.Fatal("Error field not set")
	}
}

func TestErrorNotLostOnFieldNotNamedError(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("omg", errors.New("wild walrus")))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["omg"] != "wild walrus" {
		t.Fatal("Error field not set")
	}
}

func TestFieldClashWithTime(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("timestamp", "right now!"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["fields.timestamp"] != "right now!" {
		t.Fatal("fields.timestamp not set to original time field")
	}

	cur := &time.Time{}
	if entry["timestamp"].(map[string]interface{})["seconds"] != float64(cur.Unix()) {
		t.Fatalf("timestamp field not set to current time (%d), was:  %+v", cur.Unix(), entry["timestamp"])
	}
}

func TestFieldClashWithMsg(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("message", "something"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["fields.message"] != "something" {
		t.Fatal("fields.message not set to original message field")
	}
}

func TestFieldClashWithLevel(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("level", "something"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["fields.level"] != "something" {
		t.Fatal("fields.level not set to original level field")
	}
}

func TestJSONEntryEndsWithNewline(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("level", "something"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	if b[len(b)-1] != '\n' {
		t.Fatal("Expected JSON log entry to end with a newline")
	}
}

func TestFieldDoesNotClashWithCaller(t *testing.T) {
	log.SetReportCaller(false)
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("func", "howdy pardner"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["func"] != "howdy pardner" {
		t.Fatal("func field replaced when ReportCaller=false")
	}
}

func TestFieldClashWithCaller(t *testing.T) {
	log.SetReportCaller(true)
	formatter := &Formatter{}
	e := log.WithField("logging.googleapis.com/sourceLocation", "howdy pardner")
	e.Caller = &runtime.Frame{Function: "somefunc"}
	b, err := formatter.Format(e)
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}

	entry := make(map[string]interface{})
	err = json.Unmarshal(b, &entry)
	if err != nil {
		t.Fatal("Unable to unmarshal formatted entry: ", err)
	}

	if entry["fields.logging.googleapis.com/sourceLocation"] != "howdy pardner" {
		t.Fatalf("fields.logging.googleapis.com/sourceLocation not set to original sourceLocation field when ReportCaller=true (got '%s')",
			entry["fields.logging.googleapis.com/sourceLocation"])
	}

	if entry["logging.googleapis.com/sourceLocation"].(map[string]interface{})["function"] != "somefunc" {
		t.Fatalf("logging.googleapis.com/sourceLocation.function not set as expected when ReportCaller=true (got '%s')",
			entry["logging.googleapis.com/sourceLocation"].(map[string]interface{})["function"])
	}

	log.SetReportCaller(false) // return to default value
}

func TestJSONDisableTimestamp(t *testing.T) {
	formatter := &Formatter{
		DisableTimestamp: true,
	}

	b, err := formatter.Format(log.WithField("level", "something"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}
	s := string(b)
	if strings.Contains(s, log.FieldKeyTime) {
		t.Error("Did not prevent timestamp", s)
	}
}

func TestJSONEnableTimestamp(t *testing.T) {
	formatter := &Formatter{}

	b, err := formatter.Format(log.WithField("level", "something"))
	if err != nil {
		t.Fatal("Unable to format entry: ", err)
	}
	s := string(b)
	if !strings.Contains(s, log.FieldKeyTime) {
		t.Error("Timestamp not present", s)
	}
}
