package plugin

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type Event struct {
	Timestamp time.Time
	Record    map[interface{}]interface{}
	Tag       string
}

// LogLevelKey returns the value associated with the input key from the record map, or an error if the key is not found.
func (e *Event) LogLevelKey(record map[interface{}]interface{}, logLevelKey string) (interface{}, error) {
	for key, val := range record {
		var currentKey string
		switch t := key.(type) {
		case []byte:
			currentKey = string(t)
		case string:
			currentKey = t
		default:
			log.Debugf("[go plugin]: Unable to determine type of key %v\n", t)
			continue
		}

		if logLevelKey == currentKey {
			return val, nil
		}
	}

	return nil, fmt.Errorf("failed to find key %s specified by log_level_key option in log record: %v", logLevelKey, record)
}
