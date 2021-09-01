package plugin

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

type GRPCLogSenderTestSuite struct {
	suite.Suite
	config OutputPluginConfig
}

func TestGRPCLogSenderSuite(t *testing.T) {
	suite.Run(t, new(GRPCLogSenderTestSuite))
}

func (s *GRPCLogSenderTestSuite) SetupSuite() {
	s.config = OutputPluginConfig{
		PluginInstanceId:   0,
		EndpointUrl:        "test_url",
		LogGroupId:         "test_log_group_id",
		FolderId:           "test_folder_id",
		ResourceId:         "test_resource_id",
		ResourceType:       "test_resource_type",
		PrivateKeyFilePath: "testdata/private.json",
		LogLevelKey:        "log_level",
	}
}

func (s *GRPCLogSenderTestSuite) TestGrpcLogSender_Send() {
	logLevelVal := "DEBUG"
	eventCount := 5

	s.config.LogGroupId = os.Getenv("LOG_GROUP_ID")
	s.config.ResourceId = "test_resource"
	s.config.ResourceType = "test_logs_type"
	s.config.KeyID = os.Getenv("KEY_ID")
	s.config.ServiceAccountID = os.Getenv("SERVICE_ACCOUNT_ID")
	s.config.PrivateKeyFilePath = os.Getenv("PRIVATE_KEY_FILE_PATH")
	s.config.FolderId = ""

	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	sender, err := NewGRPCLogSender(ctx, s.config)
	require.NoError(s.T(), err)

	var events []*Event
	for i := 0; i < eventCount; i++ {
		m := map[interface{}]interface{}{
			s.config.LogLevelKey: logLevelVal,
			"message":            "test_message",
			"key1":               "new_value1",
			"key2":               "new_value2",
			"key3":               "new_value3",
		}
		event := &Event{
			Timestamp: time.Now(),
			Record:    m,
			Tag:       "tt",
		}
		events = append(events, event)
	}

	err = sender.Send(events)
	require.NoError(s.T(), err)
}
