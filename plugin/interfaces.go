package plugin

import (
	"time"
	"yandex_logging/plugin/dto"
)

// LogSender is used to send Event list to external system
type LogSender interface {

	// Send sends Event list
	Send(events []*Event) error

	// getToken return either new created token or existed token if it has not been expired
	getToken() (string, error)

	// doRequest does request
	doRequest(reqModel *dto.YCLogRecordRequestModel) error
}

// OutputPlugin is the interface for output plugin
type OutputPlugin interface {
	// Flush flushes accumulated events and clear existed slice of Event list
	Flush() error

	// AddEvent adds new Event to the Event list
	AddEvent(event *Event) int

	// GetPluginInstanceID return ID of the plugin instance
	GetPluginInstanceID() int
}

type requestHandler func(reqModel *dto.YCLogRecordRequestModel) error

type authToken struct {
	token     string
	expiresAt time.Time
}
