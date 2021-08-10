package plugin

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"
	"yandex_logging/plugin/dto"
)

var ps256WithSaltLengthEqualsHash = &jwt.SigningMethodRSAPSS{
	SigningMethodRSA: jwt.SigningMethodPS256.SigningMethodRSA,
	Options: &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	},
}

type authToken struct {
	token     string
	expiresAt time.Time
}
type requestHandler func(reqModel *dto.YCLogRecordRequestModel) error

type yandexCloudHTTPClient struct {
	requestTimeout     time.Duration
	tokenLifetime      time.Duration
	authToken        authToken
	doRequestHandler requestHandler
	config           OutputPluginConfig
}

func NewYandexCloudHTTPClient(config OutputPluginConfig) *yandexCloudHTTPClient {
	cl := &yandexCloudHTTPClient{
		config: config,
		requestTimeout:     time.Second * 5,
		tokenLifetime:      time.Minute * 5,
	}
	cl.doRequestHandler = cl.doRequest
	return cl
}

func (y *yandexCloudHTTPClient) Send(events []*Event) error {

	var entries []*dto.YCLogRecordEntry
	for _, e := range events {
		logKeyStr := "LEVEL_UNSPECIFIED"

		logKey, err := e.LogLevelKey(e.Record, y.config.LogLevelKey)
		if err != nil {
			log.Errorln(errors.Wrapf(err, "ignoring log key"))
		} else {
			logKeyCasted, ok := logKey.(string)
			if !ok {
				log.Errorf("could cast log level key")
				continue
			}
			logKeyStr = logKeyCasted
		}

		entries = append(entries, &dto.YCLogRecordEntry{
			Timestamp:   e.Timestamp,
			Level:       logKeyStr,
			JsonPayload: e.Record,
		})
	}

	reqModel := &dto.YCLogRecordRequestModel{
		Destination: dto.YCLogRecordDestination{LogGroupID: y.config.LogGroupId, FolderId: y.config.FolderId},
		Resource:    dto.YCLogRecordResource{ID: y.config.ResourceId, Type: y.config.ResourceType},
		Entries:     entries,
	}
	if err := reqModel.Validate(); err != nil{
		return err
	}

	if err := y.doRequestHandler(reqModel); err != nil {
		return err
	}

	return nil
}

func (y *yandexCloudHTTPClient) doRequest(reqModel *dto.YCLogRecordRequestModel) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	b, err := json.Marshal(reqModel)
	if err != nil {
		return errors.Wrapf(err, "unable to marshal request model")
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(y.config.EndpointUrl)
	req.Header.SetMethod(http.MethodPost)
	req.Header.SetContentType("application/json")
	req.Header.SetUserAgent(fmt.Sprintf("yandexcloud-fluent-bit-plugin (%s)", runtime.GOOS))

	token, err := y.getToken()
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.SetBody(b)

	//log.Debugln(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	err = fasthttp.DoTimeout(req, resp, y.requestTimeout)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("an error occure while sending logs to yandex cloud: status_code: %d, body:%s", resp.StatusCode(), resp.Body())
	}
	return nil
}

func (y *yandexCloudHTTPClient) getToken() (string, error) {
	if y.authToken.expiresAt.Before(time.Now()) {
		authToken, err := y.createToken()
		if err != nil {
			return "", err
		}
		return authToken.token, nil
	}
	return y.authToken.token, nil
}

func (y *yandexCloudHTTPClient) createToken() (authToken, error) {

	issuedAt := time.Now()
	expiresAt := issuedAt.Add(y.tokenLifetime)
	token := jwt.NewWithClaims(ps256WithSaltLengthEqualsHash, jwt.StandardClaims{
		Issuer:    y.config.ServiceAccountID,
		IssuedAt:  issuedAt.Unix(),
		ExpiresAt: expiresAt.Unix(),
		Audience:  "https://iam.api.cloud.yandex.net/iam/v1/tokens",
	})
	token.Header["kid"] = y.config.KeyID

	f, err := os.OpenFile(y.config.PrivateKeyFilePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return authToken{}, err
	}
	defer f.Close()

	privateKey, err := y.loadPrivateKey(f)
	if err != nil {
		return authToken{}, err
	}
	signed, err := token.SignedString(privateKey)
	if err != nil {
		return authToken{}, err
	}

	y.authToken = authToken{
		token:     signed,
		expiresAt: expiresAt,
	}
	return y.authToken, nil
}

func (y *yandexCloudHTTPClient) loadPrivateKey(r io.Reader) (*rsa.PrivateKey, error) {

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
