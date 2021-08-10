package dto

import (
	valid "github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestValidatePositive(t *testing.T) {
	entries := []*YCLogRecordEntry{
		{
			Timestamp: time.Now(),
			Level:     "DEBUG",
			JsonPayload: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	reqModel := YCLogRecordRequestModel{
		Destination: YCLogRecordDestination{
			LogGroupID: "test_log_group_id",
			FolderId:   "test_folder_id",
		},
		Resource: YCLogRecordResource{
			ID:   "test_id",
			Type: "test_type",
		},
		Entries: entries,
	}

	err := reqModel.Validate()
	assert.NoError(t, err, "reqModel should not contains any error on validation")
}

func TestValidateDestinationLogGroupIDFolderIdAreEmptyNegative(t *testing.T) {

	entries := []*YCLogRecordEntry{
		{
			Timestamp: time.Now(),
			Level:     "DEBUG",
			JsonPayload: map[interface{}]interface{}{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
	}

	reqModel := YCLogRecordRequestModel{
		Destination: YCLogRecordDestination{
			LogGroupID: "",
			FolderId:   "",
		},
		Resource: YCLogRecordResource{
			ID:   "test_id",
			Type: "test_type",
		},
		Entries: entries,
	}

	err := reqModel.Validate()
	assert.Error(t, err, "either LogGroupID or FolderId must be specified")
	errors := err.(valid.ValidationErrors)
	assert.Equal(t, 2, len(errors))
}

func TestValidateDestinationOneIsSpecified(t *testing.T) {

	destinations := []YCLogRecordDestination{
		{
			LogGroupID: "test_log_group_id",
			FolderId:   "",
		},
		{
			LogGroupID: "",
			FolderId:   "test_folder_id",
		},
	}

	for _, dest := range destinations {
		entries := []*YCLogRecordEntry{
			{
				Timestamp: time.Now(),
				Level:     "DEBUG",
				JsonPayload: map[interface{}]interface{}{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		}
		reqModel := YCLogRecordRequestModel{
			Destination: dest,
			Resource: YCLogRecordResource{
				ID:   "test_id",
				Type: "test_type",
			},
			Entries: entries,
		}

		err := reqModel.Validate()
		assert.NoError(t, err, "either LogGroupID or FolderId must be specified")
	}
}
