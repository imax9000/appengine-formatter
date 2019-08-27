package appengine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Formatter formats logs into parsable json that attempts to follow
// https://cloud.google.com/logging/docs/agent/configuration#special-fields as
// closely as possible.
// Forked from logrus JSONFormatter.
type Formatter struct {
	// DisableTimestamp allows disabling automatic timestamps in output
	DisableTimestamp bool

	// CallerPrettyfier can be set by the user to modify the content
	// of the function and file keys in the json data when ReportCaller is
	// activated. If any of the returned value is the empty string the
	// corresponding key will be removed from json fields.
	// NOTE: Unlike with default formatters, you don't need to include line
	// number in `file` string. It will be automatically added as a separate
	// field.
	CallerPrettyfier func(*runtime.Frame) (function string, file string)

	// PrettyPrint will indent all json logs
	PrettyPrint bool
}

func stackdriverLevel(l log.Level) string {
	switch l {
	case log.PanicLevel, log.FatalLevel:
		return "CRITICAL"
	case log.ErrorLevel:
		return "ERROR"
	case log.WarnLevel:
		return "WARNING"
	case log.InfoLevel:
		return "INFO"
	case log.DebugLevel, log.TraceLevel:
		return "DEBUG"
	default:
		return "DEFAULT"
	}
}

// Format renders a single log entry
func (f *Formatter) Format(entry *log.Entry) ([]byte, error) {
	data := make(log.Fields, len(entry.Data)+4)

	if !f.DisableTimestamp {
		data["timestamp"] = map[string]interface{}{
			"seconds": entry.Time.Unix(),
			"nanos":   entry.Time.Nanosecond(),
		}
	}
	data["message"] = entry.Message
	data["severity"] = stackdriverLevel(entry.Level)
	data["level"] = entry.Level.String()
	if entry.HasCaller() {
		l := map[string]interface{}{}
		funcVal := entry.Caller.Function
		fileVal := entry.Caller.File
		if f.CallerPrettyfier != nil {
			funcVal, fileVal = f.CallerPrettyfier(entry.Caller)
		}
		if funcVal != "" {
			l["function"] = funcVal
		}
		if fileVal != "" {
			l["file"] = fileVal
			l["line"] = entry.Caller.Line
		}
		data["logging.googleapis.com/sourceLocation"] = l
	}

	for k, v := range entry.Data {
		if _, set := data[k]; set {
			k = "fields." + k
		}
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			// https://github.com/sirupsen/logrus/issues/137
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	encoder := json.NewEncoder(b)
	if f.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return b.Bytes(), nil
}
