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

// PopLogLevel returns the value associated with the input key from the record map, or an error if the key is not found.
func (e *Event) PopLogLevel(record map[interface{}]interface{}, logLevelKey string) (string, error) {
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
			logKeyCasted, ok := val.(string)
			if !ok {
				return "", fmt.Errorf("could cast log level key")
			}
			delete(record, key)
			return logKeyCasted, nil
		}
	}

	return "", fmt.Errorf("failed to find key %s specified by log_level_key option in log record: %v", logLevelKey, record)
}

func (e *Event) PopMessageKey(record map[interface{}]interface{}, messageKey string) (string, error) {

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

		if messageKey == currentKey {
			v, ok := val.(string)
			if ok {
				delete(record, key)
				return v, nil
			}
			return "", fmt.Errorf("could not cast message key to string")
		}
	}

	return "", fmt.Errorf("failed to find key %s specified by log_level_key option in log record: %v", messageKey, record)
}
