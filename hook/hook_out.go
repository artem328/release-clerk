package hook

import (
	"encoding/json"
	"fmt"
)

var outputPayloadRegistry = make(map[string]func() OutputPayload)

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

type Output struct {
	Type    string
	Logs    []Log
	Error   *Error        `json:",omitempty"`
	Payload OutputPayload `json:",omitempty"`
}

func (o Output) Err() error {
	if o.Error == nil {
		return nil
	}
	return o.Error
}

func (o *Output) UnmarshalJSON(data []byte) error {
	var out struct {
		Type    string
		Logs    []Log
		Error   *Error
		Payload json.RawMessage
	}

	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	constr, ok := outputPayloadRegistry[out.Type]
	if !ok {
		return fmt.Errorf("hook `%s` is not registered", out.Type)
	}

	o.Type = out.Type
	o.Logs = out.Logs
	o.Error = out.Error
	o.Payload = constr()

	if len(out.Payload) == 0 {
		return nil
	}

	return json.Unmarshal(out.Payload, &o.Payload)
}

type OutputPayload interface {
	isOutputPayload()
}

func AsOutputPayload[P OutputPayload](p OutputPayload) (P, bool) {
	pp, ok := p.(P)

	return pp, ok
}
