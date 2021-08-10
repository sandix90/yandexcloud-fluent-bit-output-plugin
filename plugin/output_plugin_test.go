package plugin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func Test_OutputPlugin_Flush(t *testing.T) {
	mockLogSender := &MockLogSender{}
	mockLogSender.On("Send", mock.AnythingOfType("[]*plugin.Event")).Return(nil)
	plugin := NewYandexCloudOutputPlugin(OutputPluginConfig{}, mockLogSender)

	assert.Equal(t, 0, len(plugin.events), "There are must be 0 event inside")
	eventsQuantity := 5
	for i := 0; i < eventsQuantity; i++ {
		plugin.AddEvent(&Event{
			Timestamp: time.Now(),
			Record:    map[interface{}]interface{}{"key1": "val1", "key2": "val2"},
			Tag:       "test_tag",
		})
	}

	assert.Equal(t, 5, len(plugin.events), "There are must be 5 event inside")
	err := plugin.Flush()
	assert.NoError(t, err, "Flush method should not have any errors here")
	assert.Equal(t, 0, len(plugin.events), "There are must be 0 event inside")
}

func Test_OutputPlugin_Flush_With_Error(t *testing.T) {
	mockLogSender := &MockLogSender{}
	mockLogSender.On("Send", mock.AnythingOfType("[]*plugin.Event")).Return(fmt.Errorf("some send error"))
	plugin := NewYandexCloudOutputPlugin(OutputPluginConfig{}, mockLogSender)

	assert.Equal(t, 0, len(plugin.events), "There are must be 0 event inside")
	eventsQuantity := 5
	for i := 0; i < eventsQuantity; i++ {
		plugin.AddEvent(&Event{
			Timestamp: time.Now(),
			Record:    map[interface{}]interface{}{"key1": "val1", "key2": "val2"},
			Tag:       "test_tag",
		})
	}

	assert.Equal(t, 5, len(plugin.events), "There are must be 5 event inside")
	err := plugin.Flush()
	assert.Error(t, err, "err should not be nil in this place")
	assert.Equal(t, 5, len(plugin.events), "There are must be 5 event inside")
}

func Test_OutputPlugin_Different_Events_Slices(t *testing.T) {
	mockLogSender := &MockLogSender{}
	mockLogSender.On("Send", mock.AnythingOfType("[]*plugin.Event")).Return(nil)
	plugin := NewYandexCloudOutputPlugin(OutputPluginConfig{}, mockLogSender)

	eventsQuantity := 5
	for i := 0; i < eventsQuantity; i++ {
		plugin.AddEvent(&Event{
			Timestamp: time.Now(),
			Record:    map[interface{}]interface{}{"key1": "val1", "key2": "val2"},
			Tag:       "test_tag",
		})
	}
	firstSliceP := fmt.Sprintf("%p", &plugin.events)
	firstElemSliceP := fmt.Sprintf("%p", &plugin.events[0])
	err := plugin.Flush()
	assert.NoError(t, err)

	for i := 0; i < eventsQuantity; i++ {
		plugin.AddEvent(&Event{
			Timestamp: time.Now(),
			Record:    map[interface{}]interface{}{"key1": "val1", "key2": "val2"},
			Tag:       "test_tag",
		})
	}
	secondSliceP := fmt.Sprintf("%p", &plugin.events)
	secondElemSliceP := fmt.Sprintf("%p", &plugin.events[0])

	assert.Equal(t, firstSliceP, secondSliceP, "slices addresses should not be equal")
	assert.NotEqual(t, firstElemSliceP, secondElemSliceP, "slice elems addresses should not be equal")
	assert.Equal(t, 5, len(plugin.events), "There are must be 5 event inside")
}
