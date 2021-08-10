package plugin

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
	"yandex_logging/plugin/dto"
)

type HttpLogSenderTestSuite struct {
	suite.Suite
	config OutputPluginConfig
}

func TestHttpLogSenderSuite(t *testing.T) {
	suite.Run(t, new(HttpLogSenderTestSuite))
}

func (s *HttpLogSenderTestSuite) SetupSuite() {
	s.config = OutputPluginConfig{
		PluginInstanceId:   0,
		EndpointUrl:        "test_url",
		LogGroupId:         "test_log_group_id",
		FolderId:           "test_folder_id",
		ResourceId:         "test_resource_id",
		ResourceType:       "test_resource_type",
		KeyID:              "test_key_id",
		ServiceAccountID:   "test_service_account_id",
		PrivateKeyFilePath: "testdata/test_private.pem",
		LogLevelKey:        "log_level",
	}
}

func (s *HttpLogSenderTestSuite) Test_PrivateKeyDoesntExist() {
	s.config.PrivateKeyFilePath = "not_existed_test_private.pem"
	client := &yandexCloudHTTPClient{config: s.config}

	_, err := client.getToken()
	assert.True(s.T(), errors.Is(err, os.ErrNotExist))
}

func (s *HttpLogSenderTestSuite) Test_GetTokenEqualsWithinExpiresTime() {
	client := &yandexCloudHTTPClient{
		config:        s.config,
		tokenLifetime: time.Minute * 5,
	}

	token, err := client.getToken()
	assert.NoError(s.T(), err)

	token2, err := client.getToken()
	assert.NoError(s.T(), err)

	assert.Equal(s.T(), token, token2)
}

func (s *HttpLogSenderTestSuite) Test_GetTokenNotEqualsLifetimeOutOfDate() {
	client := &yandexCloudHTTPClient{
		config:        s.config,
		tokenLifetime: time.Second * 1,
	}

	token, err := client.getToken()
	assert.NoError(s.T(), err)

	time.Sleep(2 * time.Second)
	token2, err := client.getToken()
	assert.NoError(s.T(), err)

	assert.NotEqual(s.T(), token, token2)
}

func (s *HttpLogSenderTestSuite) Test_FindLogLevelValue() {
	logLevelVal := "TEST_DEBUG"
	eventCount := 5
	client := &yandexCloudHTTPClient{
		config:         s.config,
		tokenLifetime:  time.Second * 1,
		requestTimeout: time.Second * 5,
	}
	client.doRequestHandler = func(reqModel *dto.YCLogRecordRequestModel) error {
		assert.Equal(s.T(), eventCount, len(reqModel.Entries))
		assert.Equal(s.T(), logLevelVal, reqModel.Entries[0].Level)
		return nil
	}

	var events []*Event
	for i := 0; i < eventCount; i++ {
		event := &Event{
			Timestamp: time.Now(),
			Record: map[interface{}]interface{}{
				s.config.LogLevelKey: logLevelVal,
				"key1":               "value1",
				"key2":               "value2",
				"key3":               "value3",
			},
		}

		events = append(events, event)
	}
	err := client.Send(events)
	assert.NoError(s.T(), err)
}
