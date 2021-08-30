package plugin

import (
	fluentbit "github.com/fluent/fluent-bit-go/output"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"unsafe"
)

var ErrFieldRequired = errors.New("Field required")
var ErrOneOfFieldsRequired = errors.New("One of the given fields are required")

type OutputPluginConfig struct {
	PluginInstanceId   int
	EndpointUrl        string
	LogGroupId         string
	FolderId           string
	ResourceId         string
	ResourceType       string
	KeyID              string
	ServiceAccountID   string
	PrivateKeyFilePath string
	LogLevelKey        string
}

func NewOutputPluginConfig(ctx unsafe.Pointer, pluginID int) OutputPluginConfig {
	config := OutputPluginConfig{}
	config.PluginInstanceId = pluginID

	config.EndpointUrl = fluentbit.FLBPluginConfigKey(ctx, "endpoint_url")
	if config.EndpointUrl == "" {
		config.EndpointUrl = "ingester.logging.yandexcloud.net:443"
	}
	log.Infof("[yandexcloud %d] plugin parameter endpoint_url = `%s`", pluginID, config.EndpointUrl)

	config.LogGroupId = fluentbit.FLBPluginConfigKey(ctx, "log_group_id")
	log.Infof("[yandexcloud %d] plugin parameter log_group_id = `%s`", pluginID, config.LogGroupId)

	config.FolderId = fluentbit.FLBPluginConfigKey(ctx, "folder_id")
	log.Infof("[yandexcloud %d] plugin parameter folder_id = `%s`", pluginID, config.FolderId)

	config.ResourceId = fluentbit.FLBPluginConfigKey(ctx, "resource_id")
	log.Infof("[yandexcloud %d] plugin parameter resource_id = `%s`", pluginID, config.ResourceId)

	config.ResourceType = fluentbit.FLBPluginConfigKey(ctx, "resource_type")
	log.Infof("[yandexcloud %d] plugin parameter resource_type = `%s`", pluginID, config.ResourceType)

	config.KeyID = fluentbit.FLBPluginConfigKey(ctx, "key_id")
	log.Infof("[yandexcloud %d] plugin parameter key_id = `%s`", pluginID, config.KeyID)

	config.ServiceAccountID = fluentbit.FLBPluginConfigKey(ctx, "service_account_id")
	log.Infof("[yandexcloud %d] plugin parameter service_account_id = `%s`", pluginID, config.ServiceAccountID)

	config.PrivateKeyFilePath = fluentbit.FLBPluginConfigKey(ctx, "private_key_file_path")
	log.Infof("[yandexcloud %d] plugin parameter private_key_file_path = `%s`", pluginID, config.PrivateKeyFilePath)

	config.LogLevelKey = fluentbit.FLBPluginConfigKey(ctx, "log_level_key")
	if config.LogLevelKey == "" {
		config.LogLevelKey = "level"
	}
	log.Infof("[yandexcloud %d] plugin parameter log_level_key = `%s`", pluginID, config.LogLevelKey)

	return config
}

func (config OutputPluginConfig) Validate() error {

	if config.LogGroupId == "" && config.FolderId == "" {
		return errors.Wrap(ErrOneOfFieldsRequired, "log_group_id or folder_id")
	}

	if config.ResourceId == "" {
		return errors.Wrap(ErrFieldRequired, "resource_id")
	}

	if config.ResourceType == "" {
		return errors.Wrap(ErrFieldRequired, "resource_type")
	}

	if config.KeyID == "" {
		return errors.Wrap(ErrFieldRequired, "key_id")
	}

	if config.ServiceAccountID == "" {
		return errors.Wrap(ErrFieldRequired, "service_account_id")
	}

	if config.PrivateKeyFilePath == "" {
		return errors.Wrap(ErrFieldRequired, "private_key_file_path")
	}

	return nil
}
