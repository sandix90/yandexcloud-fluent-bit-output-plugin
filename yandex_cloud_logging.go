package main

import (
	"C"
	fluentbit "github.com/fluent/fluent-bit-go/output"
	log "github.com/sirupsen/logrus"
	"time"
	"unsafe"
	"yandex_logging/plugin"
)

var (
	pluginInstances []plugin.OutputPlugin
)

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return fluentbit.FLBPluginRegister(ctx, "yandex_cloud", "Yandex Cloud Logging Fluent Bit Plugin.")
}

//export FLBPluginInit
func FLBPluginInit(ctx unsafe.Pointer) int {

	log.Debug("Yandex Cloud Logging plugin")

	err := addPluginInstance(ctx)
	if err != nil {
		log.Error(err)
		return fluentbit.FLB_ERROR
	}
	return fluentbit.FLB_OK
}

func addPluginInstance(ctx unsafe.Pointer) error {
	pluginID := len(pluginInstances)

	config := plugin.NewOutputPluginConfig(ctx, pluginID)
	err := config.Validate()
	if err != nil {
		return err
	}

	logSender := plugin.NewYandexCloudHTTPClient(config)
	pluginInstance := plugin.NewYandexCloudOutputPlugin(config, logSender)

	fluentbit.FLBPluginSetContext(ctx, pluginID)
	pluginInstances = append(pluginInstances, pluginInstance)

	return nil
}

func getPluginInstance(ctx unsafe.Pointer) plugin.OutputPlugin {
	pluginID := fluentbit.FLBPluginGetContext(ctx).(int)
	return pluginInstances[pluginID]
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	dec := fluentbit.NewDecoder(data, int(length))
	ycLogPlugin := getPluginInstance(ctx)

	fluentTag := C.GoString(tag)
	log.Debugf("[yandexcloud %d] Found logs with tag: %s", ycLogPlugin.GetPluginInstanceID(), fluentTag)

	count := 0
	for {
		ret, ts, record := fluentbit.GetRecord(dec)
		if ret != 0 {
			break
		}
		log.Printf("record: %v", record)

		var timestamp time.Time
		switch t := ts.(type) {
		case fluentbit.FLBTime:
			timestamp = ts.(fluentbit.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			timestamp = time.Now()
		}

		retCode := ycLogPlugin.AddEvent(&plugin.Event{Timestamp: timestamp, Record: record, Tag: fluentTag})
		if retCode != fluentbit.FLB_OK {
			return retCode
		}

		count++
	}

	err := ycLogPlugin.Flush()
	if err != nil {
		log.Errorln(err)
		return fluentbit.FLB_RETRY
	}

	log.Debugf("[yandexcloud %d] Processed %d events", ycLogPlugin.GetPluginInstanceID(), count)
	return fluentbit.FLB_OK

}

//export FLBPluginExit
func FLBPluginExit() int {
	return fluentbit.FLB_OK
}

func main() {
}
