package hook

import (
	"encoding/json"
	"errors"
)

var ErrInputTypeMismatch = errors.New("input type mismatch")

type Input[P InputPayload] struct {
	Type    string
	Payload P
}

func (i *Input[P]) UnmarshalJSON(data []byte) error {
	var in struct {
		Type    string
		Payload json.RawMessage
	}

	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	if i.Payload.inputPayloadType() != in.Type {
		return ErrInputTypeMismatch
	}

	i.Type = in.Type

	if err := json.Unmarshal(in.Payload, &i.Payload); err != nil {
		return err
	}

	return nil
}

type InputPayload interface {
	inputPayloadType() string
}
