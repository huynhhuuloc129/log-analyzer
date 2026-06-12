package models

import (
	"encoding/json"
	"time"
)

type LogEntry struct {
	Timestamp     time.Time `json:"@t"`
	Message       string    `json:"@mt"`
	Level         string    `json:"@l"`
	EventId       json.RawMessage       `json:"@i"`
	SourceContext string    `json:"SourceContext"`
	ProcessId     int       `json:"ProcessId"`
	ThreadId      int       `json:"ThreadId"`
	RequestId     string    `json:"RequestId"`
	RequestPath   string    `json:"RequestPath"`
	ConnectionId  string    `json:"ConnectionId"`
}