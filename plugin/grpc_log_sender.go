package plugin

import (
	"context"
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	"google.golang.org/grpc"
	grpcMetadata "google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"io/ioutil"
	"os"
	"time"
	"yandex_logging/plugin/dto"
)

type grpcLogSender struct {
	config           OutputPluginConfig
	doRequestHandler requestHandler
	authToken        authToken
	tokenLifetime    time.Duration
	requestTimeout   time.Duration
	ingestionClient  logging.LogIngestionServiceClient
}

func NewGRPCLogSender(config OutputPluginConfig) (*grpcLogSender, error) {
	conn, err := grpc.Dial(config.EndpointUrl, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "grpc connect failed")
	}
	defer conn.Close()
	c := logging.NewLogIngestionServiceClient(conn)

	sender := &grpcLogSender{
		config:          config,
		tokenLifetime:   time.Minute * 5,
		requestTimeout:  time.Second * 5,
		ingestionClient: c,
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
			Timestamp:   &timestamp.Timestamp{Seconds: int64(e.Timestamp.Second())},
			Level:       logging.LogLevel_Level(logging.LogLevel_Level_value[e.Level]),
			Message:     e.Message,
			JsonPayload: nStruct,
		}

		wEntries = append(wEntries, we)
	}
	wr.SetEntries(wEntries)

	token, err := g.getToken()
	if err != nil {
		return err
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	ctx = grpcMetadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	defer cancelFn()

	response, err := g.ingestionClient.Write(ctx, &wr, nil)
	if err != nil {
		return err
	}
	log.Infoln(response)
	return nil
}

func (g *grpcLogSender) getToken() (string, error) {
	if g.authToken.expiresAt.Before(time.Now()) {
		authToken, err := g.createToken()
		if err != nil {
			return "", err
		}
		return authToken.token, nil
	}
	return g.authToken.token, nil
}

func (g *grpcLogSender) createToken() (authToken, error) {

	issuedAt := time.Now()
	expiresAt := issuedAt.Add(g.tokenLifetime)
	token := jwt.NewWithClaims(ps256WithSaltLengthEqualsHash, jwt.StandardClaims{
		Issuer:    g.config.ServiceAccountID,
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: expiresAt.Unix(),
		Audience:  "https://iam.api.cloud.yandex.net/iam/v1/tokens",
	})
	token.Header["kid"] = g.config.KeyID

	f, err := os.OpenFile(g.config.PrivateKeyFilePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return authToken{}, err
	}
	defer f.Close()

	privateKey, err := g.loadPrivateKey(f)
	if err != nil {
		return authToken{}, err
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return authToken{}, err
	}

	g.authToken = authToken{
		token:     signed,
		expiresAt: expiresAt,
	}
	return g.authToken, nil
}

func (g *grpcLogSender) loadPrivateKey(r io.Reader) (*rsa.PrivateKey, error) {

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	rsaPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return nil, err
	}
	return rsaPrivateKey, nil
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
