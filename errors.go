package clhplugin

import (
	"errors"
	"fmt"
)

var (
	ErrClientClosed    = errors.New("client is closed")
	ErrNotConnected    = errors.New("client is not connected")
	ErrInvalidManifest = errors.New("invalid plugin manifest")
)

type RemoteError struct {
	Topic         EnvelopeTopic
	Code          string
	Message       string
	CorrelationID string
}

func (e *RemoteError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("remote error topic=%d code=%s message=%s correlation_id=%s", e.Topic, e.Code, e.Message, e.CorrelationID)
}
