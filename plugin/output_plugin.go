package plugin

import (
	fluentbit "github.com/fluent/fluent-bit-go/output"
)

type ycOutputPlugin struct {
	pluginInstanceID int
	logSender        LogSender
	events           []*Event
}

func NewYandexCloudOutputPlugin(config OutputPluginConfig, logSender LogSender) *ycOutputPlugin {
	return &ycOutputPlugin{
		pluginInstanceID: config.PluginInstanceId,
		logSender:        logSender,
	}
}

func (p *ycOutputPlugin) Flush() error {
	err := p.logSender.Send(p.events)
	if err != nil {
		return err
	}
	p.events = nil
	return nil
}

func (p *ycOutputPlugin) AddEvent(event *Event) int {
	p.events = append(p.events, event)
	return fluentbit.FLB_OK
}

func (p *ycOutputPlugin) GetPluginInstanceID() int {
	return p.pluginInstanceID
}
