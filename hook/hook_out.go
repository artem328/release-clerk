package hook

import (
	"encoding/json"
	"errors"
)

var ErrOutputTypeMismatch = errors.New("output type mismatch")

type Error struct {
	Message string
}

func (e Error) Error() string {
	return e.Message
}

type Log struct {
	Message string
	Debug   bool
}

type Output[P OutputPayload] struct {
	Type    string
	Logs    []Log
	Error   *Error `json:",omitempty"`
	Payload *P     `json:",omitempty"`
}

func (o Output[P]) Err() error {
	if o.Error == nil {
		return nil
	}
	return o.Error
}

func (o *Output[P]) UnmarshalJSON(data []byte) error {
	var out struct {
		Type    string
		Logs    []Log
		Error   *Error
		Payload json.RawMessage
	}

	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	p := new(P)
	if (*p).outputPayloadType() != out.Type {
		return ErrOutputTypeMismatch
	}

	o.Type = out.Type
	o.Logs = out.Logs
	o.Error = out.Error

	if len(out.Payload) == 0 {
		return nil
	}

	o.Payload = p

	return json.Unmarshal(out.Payload, &o.Payload)
}

type OutputPayload interface {
	outputPayloadType() string
}
