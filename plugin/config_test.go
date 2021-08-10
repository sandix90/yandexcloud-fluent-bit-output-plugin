package plugin

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Config_Validate(t *testing.T) {
	errorsSeq := []error{ErrOneOfFieldsRequired, ErrFieldRequired, ErrFieldRequired,
		ErrFieldRequired, ErrFieldRequired, ErrFieldRequired,
	}
	logLevelKey := "log_level"
	configs := []OutputPluginConfig{
		{
			0,
			"",
			"", // <-- testing this and
			"",   // <-- this fields
			"test_resource_id",
			"test_resource_type",
			"test_key_id",
			"test_service_account",
			"test_private_ket_path",
			logLevelKey,
		},
		{
			0,
			"",
			"test_log_group",
			"test_resource_type",
			"", // <-- testing this field
			"test_resource_type",
			"test_key_id",
			"test_service_account",
			"test_private_ket_path",
			logLevelKey,
		},
		{
			0,
			"",
			"test_log_group_id",
			"test_folder_id",
			"test_resource_id",
			"", // <-- testing this field
			"test_key_id",
			"test_service_account",
			"test_private_ket_path",
			logLevelKey,
		},
		{
			0,
			"",
			"test_log_group_id",
			"test_folder_id",
			"test_resource_id",
			"test_resource_type",
			"", // <-- testing this field
			"test_service_account",
			"test_private_ket_path",
			logLevelKey,
		},
		{
			0,
			"",
			"test_log_group_id",
			"test_folder_id",
			"test_resource_id",
			"test_resource_type",
			"test_key_id",
			"", // <-- testing this field
			"test_private_ket_path",
			logLevelKey,
		},
		{
			0,
			"",
			"test_log_group_id",
			"test_folder_id",
			"test_resource_id",
			"test_resource_type",
			"test_key_id",
			"test_service_account",
			"", // <-- testing this field
			logLevelKey,
		},
	}

	for idx, config := range configs {
		err := config.Validate()
		assert.True(t, errors.Is(err, errorsSeq[idx]), "should have error here")
	}
}
