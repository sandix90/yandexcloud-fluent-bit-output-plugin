package dto

import (
	"github.com/go-playground/validator/v10"
	"time"
)

var validate = validator.New()

type YCLogRecordDestination struct {
	LogGroupID string `json:"logGroupId" validate:"required_without=FolderId"`
	FolderId   string `json:"folderId" validate:"required_without=LogGroupID"`
}

type YCLogRecordResource struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type YCLogRecordEntry struct {
	Timestamp   time.Time                   `json:"timestamp" validate:"required"`
	Level       string                      `json:"level" validate:"required"`
	JsonPayload map[interface{}]interface{} `json:"jsonPayload" validate:"required"`
}

type YCLogRecordRequestModel struct {
	Destination YCLogRecordDestination `json:"destination" validate:"required,dive"`
	Resource    YCLogRecordResource    `json:"resource"`
	Entries     []*YCLogRecordEntry    `json:"entries" validate:"required,dive"`
}

func (d *YCLogRecordRequestModel) Validate() error {
	return validate.Struct(d)
}
