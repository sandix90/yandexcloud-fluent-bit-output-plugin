package plugin

import (
	"context"
	"encoding/json"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"io/ioutil"
	"time"
	"yandex_logging/plugin/dto"
)

type iamKey struct {
	ServiceAccountID string `json:"service_account_id"`
	ID               string `json:"id"`
	PrivateKey       string `json:"private_key"`
}

type grpcLogSender struct {
	config           OutputPluginConfig
	doRequestHandler requestHandler
	authToken        authToken
	tokenLifetime    time.Duration
	requestTimeout   time.Duration
	parentCtx        context.Context
	sdk              *ycsdk.SDK
}

func NewGRPCLogSender(ctx context.Context, config OutputPluginConfig) (*grpcLogSender, error) {

	privateBuffer, err := ioutil.ReadFile(config.PrivateKeyFilePath)
	if err != nil {
		return nil, err
	}

	ikey := &iamKey{
		ServiceAccountID: config.ServiceAccountID,
		ID:               config.KeyID,
		PrivateKey:       string(privateBuffer),
	}

	iKeyBytes, err := json.Marshal(ikey)
	if err != nil {
		return nil, err
	}

	k, err := iamkey.ReadFromJSONBytes(iKeyBytes)
	if err != nil {
		return nil, err
	}

	creds, err := ycsdk.ServiceAccountKey(k)
	if err != nil {
		return nil, err
	}

	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: creds,
	})
	if err != nil {
		return nil, err
	}

	sender := &grpcLogSender{
		config:         config,
		tokenLifetime:  time.Minute * 5,
		requestTimeout: time.Second * 5,
		parentCtx:      ctx,
		sdk:            sdk,
	}
	sender.doRequestHandler = sender.doRequest
	return sender, nil
}

func (g *grpcLogSender) Send(events []*Event) error {
	var entries []*dto.YCLogRecordEntry
	for _, e := range events {
		logLevelVal := "LEVEL_UNSPECIFIED"

		logLevelVal, err := e.PopLogLevel(e.Record, g.config.LogLevelKey)
		if err != nil {
			log.Errorln(err)
		}

		message, err := e.PopMessageKey(e.Record, "message")
		if err != nil {
			log.Errorln(err)
		}

		entries = append(entries, &dto.YCLogRecordEntry{
			Timestamp:   e.Timestamp,
			Level:       logLevelVal,
			JsonPayload: e.Record,
			Message:     message,
		})
	}

	reqModel := &dto.YCLogRecordRequestModel{
		Destination: dto.YCLogRecordDestination{LogGroupID: g.config.LogGroupId, FolderId: g.config.FolderId},
		Resource:    dto.YCLogRecordResource{ID: g.config.ResourceId, Type: g.config.ResourceType},
		Entries:     entries,
	}
	if err := reqModel.Validate(); err != nil {
		return err
	}

	if err := g.doRequestHandler(reqModel); err != nil {
		return err
	}

	return nil
}

func (g *grpcLogSender) doRequest(reqModel *dto.YCLogRecordRequestModel) error {

	wr := logging.WriteRequest{}

	var destination logging.Destination
	if reqModel.Destination.FolderId != "" {
		destination.Destination = &logging.Destination_FolderId{FolderId: reqModel.Destination.FolderId}
	} else {
		destination.Destination = &logging.Destination_LogGroupId{LogGroupId: reqModel.Destination.LogGroupID}
	}
	wr.SetDestination(&destination)

	wResource := &logging.LogEntryResource{
		Type: reqModel.Resource.Type,
		Id:   reqModel.Resource.ID,
	}
	wr.SetResource(wResource)

	var wEntries []*logging.IncomingLogEntry
	for _, e := range reqModel.Entries {
		nStruct, err := structpb.NewStruct(g.convertMap(e.JsonPayload))
		if err != nil {
			log.Errorln(errors.Wrapf(err, "cannot prepare struct for incomingLogEntry"))
			continue
		}
		we := &logging.IncomingLogEntry{
			Timestamp:   &timestamp.Timestamp{Seconds: e.Timestamp.Unix()},
			Level:       logging.LogLevel_Level(logging.LogLevel_Level_value[e.Level]),
			Message:     e.Message,
			JsonPayload: nStruct,
		}

		wEntries = append(wEntries, we)
	}
	wr.SetEntries(wEntries)

	ctx, cancelFn := context.WithTimeout(g.parentCtx, g.requestTimeout)
	defer cancelFn()
	response, err := g.sdk.LogIngestion().LogIngestion().Write(ctx, &wr, grpc.EmptyCallOption{})
	if err != nil {
		return err
	}
	log.Infoln(response)
	return nil
}

func (g *grpcLogSender) getToken() (string, error) {
	return "", nil
}

func (g *grpcLogSender) convertMap(m map[interface{}]interface{}) map[string]interface{} {
	newM := make(map[string]interface{})

	for k, v := range m {
		if s, ok := k.(string); ok {
			newM[s] = v
		}
	}
	return newM
}
