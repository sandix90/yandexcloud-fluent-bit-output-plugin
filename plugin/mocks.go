package plugin

import (
	"github.com/stretchr/testify/mock"
	"yandex_logging/plugin/dto"
)

var _ OutputPlugin = (*MockOutputPlugin)(nil)

type MockOutputPlugin struct {
	mock.Mock
}

func (m *MockOutputPlugin) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOutputPlugin) AddEvent(event *Event) int {
	args := m.Called(event)
	return args.Get(0).(int)
}

func (m *MockOutputPlugin) GetPluginInstanceID() int {
	args := m.Called()
	return args.Get(0).(int)
}

var _ LogSender = (*MockLogSender)(nil)

type MockLogSender struct {
	mock.Mock
}

func (m *MockLogSender) Send(events []*Event) error {
	args := m.Called(events)
	return args.Error(0)
}

func (m *MockLogSender) getToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockLogSender) doRequest(reqModel *dto.YCLogRecordRequestModel) error {
	args := m.Called(reqModel)
	return args.Error(0)
}
