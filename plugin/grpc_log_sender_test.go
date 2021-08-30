package plugin

import (
	"context"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
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
		KeyID:              "test_key_id",
		ServiceAccountID:   "test_service_account_id",
		PrivateKeyFilePath: "testdata/test_private.pem",
		LogLevelKey:        "log_level",
	}
}

func (s *GRPCLogSenderTestSuite) Test_FindLogLevelValue() {
	logLevelVal := "TEST_DEBUG"
	eventCount := 5

	/*log_group_id e23v58hh1tu4e536a4hl
	resource_id pivoman_test_resource
	resource_type test_logs_type
	endpoint_url ingester.logging.yandexcloud.net
	key_id ajejnbhfmf3rs29lj6vl
	service_account_id ajeac8j4a6cilovgesoo
	private_key_file_path /fluent-bit/etc/private.pem*/
	s.config.LogGroupId = "e23v58hh1tu4e536a4hl"
	s.config.EndpointUrl = "ingester.logging.yandexcloud.net:443"
	//s.config.EndpointUrl = "api.cloud.yandex.net:443"
	s.config.ResourceId = "pivoman_test_resource"
	s.config.ResourceType = "test_logs_type"
	s.config.KeyID = "ajejnbhfmf3rs29lj6vl"
	s.config.ServiceAccountID = "ajeac8j4a6cilovgesoo"
	s.config.PrivateKeyFilePath = "fluent-bit/etc/private.pem"

	ctx, cancelConn := context.WithTimeout(context.Background(), time.Second*55)
	defer cancelConn()

	//iamkey.ReadFromJSONBytes()
	k, err := iamkey.ReadFromJSONFile("./testdata/service_account.json")
	require.NoError(s.T(), err)

	creds, err := ycsdk.ServiceAccountKey(k)
	require.NoError(s.T(), err)

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: creds,
	})
	require.NoError(s.T(), err)

	//t, err  := sdk.CreateIAMTokenForServiceAccount(ctx, "ajeac8j4a6cilovgesoo")
	//require.NoError(s.T(), err)
	//log.Infoln(t)

	//list, err := sdk.Logging().LogGroup().List(ctx, nil, grpc.WithBlock())
	//conn, err := grpc.DialContext(ctx, s.config.EndpointUrl, grpc.WithBlock(), grpc.WithInsecure())
	//require.NoError(s.T(), err)
	//defer conn.Close()
	//c := logging.NewLogIngestionServiceClient(conn)
	//
	//client := &grpcLogSender{
	//	config:          s.config,
	//	tokenLifetime:   time.Second * 1,
	//	requestTimeout:  time.Second * 5,
	//	ingestionClient: c,
	//}
	//client.doRequestHandler = func(reqModel *dto.YCLogRecordRequestModel) error {
	//	assert.Equal(s.T(), eventCount, len(reqModel.Entries))
	//	assert.Equal(s.T(), logLevelVal, reqModel.Entries[0].Level)
	//	return nil
	//}

	var events []*logging.IncomingLogEntry
	for i := 0; i < eventCount; i++ {
		m := map[string]interface{}{
			s.config.LogLevelKey: logLevelVal,
			"key1":               "value1",
			"key2":               "value2",
			"key3":               "value3",
		}
		newStruct, err := structpb.NewStruct(m)
		require.NoError(s.T(), err)
		//require.NoError(s.T(), err)

		we := &logging.IncomingLogEntry{
			Timestamp:   &timestamp.Timestamp{Seconds: time.Now().Unix()},
			Level:       logging.LogLevel_INFO,
			Message:     "test message",
			JsonPayload: newStruct,
		}
		events = append(events, we)
	}

	wr := logging.WriteRequest{
		Entries: events,
	}
	wr.Destination = &logging.Destination{Destination: &logging.Destination_FolderId{FolderId: "adf"}}
	wResource := &logging.LogEntryResource{
		Type: "test_type",
		Id:   "test_id",
	}
	wr.SetResource(wResource)
	wl := logging.ListLogGroupsRequest{
		FolderId: "b1giufb956v3ib6oi107",
	}

	response, err := sdk.Logging().LogGroup().List(ctx, &wl, grpc.EmptyCallOption{})
	//response, err := sdk.LogIngestion().LogIngestion().Write(ctx, &wr, grpc.EmptyCallOption{})
	require.NoError(s.T(), err)
	log.Infoln(response)

	//err = client.Send(events)
	//assert.NoError(s.T(), err)
}
