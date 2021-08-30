package plugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEvent_GetMessageKey(t *testing.T) {
	t.Run("pop_up_message", func(t *testing.T) {
		messageKey := "message"
		record := map[interface{}]interface{}{"key1": "val1", "key2": "val2", messageKey: "test_message"}
		e := Event{
			Timestamp: time.Now(),
			Record:    record,
			Tag:       "test_tag",
		}

		_, ok := record[messageKey]
		assert.True(t, ok)

		m, err := e.PopMessageKey(e.Record, "message")
		assert.NoError(t, err)
		assert.Equal(t, "test_message", m)

		_, ok = record[messageKey]
		assert.False(t, ok)
	})

	t.Run("absence_of_message", func(t *testing.T) {
		record := map[interface{}]interface{}{"key1": "val1", "key2": "val2"}
		e := Event{
			Timestamp: time.Now(),
			Record:    record,
			Tag:       "test_tag",
		}

		_, err := e.PopMessageKey(e.Record, "message")
		assert.Error(t, err)

	})

}

func TestEvent_GetLogLevel(t *testing.T) {
	t.Run("pop_up_log_level", func(t *testing.T) {
		logLevelKey := "log_level"
		record := map[interface{}]interface{}{"key1": "val1", "key2": "val2", logLevelKey: "DEBUG"}
		e := Event{
			Timestamp: time.Now(),
			Record:    record,
			Tag:       "test_tag",
		}

		_, ok := record[logLevelKey]
		assert.True(t, ok)

		m, err := e.PopLogLevel(e.Record, logLevelKey)
		assert.NoError(t, err)
		assert.Equal(t, "DEBUG", m)

		_, ok = record[logLevelKey]
		assert.False(t, ok)
	})

	t.Run("absence_of_log_level", func(t *testing.T) {
		record := map[interface{}]interface{}{"key1": "val1", "key2": "val2"}
		e := Event{
			Timestamp: time.Now(),
			Record:    record,
			Tag:       "test_tag",
		}

		_, err := e.PopLogLevel(e.Record, "DEBUG")
		assert.Error(t, err)

	})

}
