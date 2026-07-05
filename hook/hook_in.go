package hook

import (
	"encoding/json"
	"fmt"
)

var inputPayloadRegistry = make(map[string]func() InputPayload)

type Input struct {
	Type    string
	Payload InputPayload
}

func (i *Input) UnmarshalJSON(data []byte) error {
	var in struct {
		Type    string
		Payload json.RawMessage
	}

	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	constr, ok := inputPayloadRegistry[in.Type]
	if !ok {
		return fmt.Errorf("hook `%s` is not registered", in.Type)
	}

	i.Type = in.Type
	i.Payload = constr()

	if err := json.Unmarshal(in.Payload, &i.Payload); err != nil {
		return err
	}

	return nil
}

type InputPayload interface {
	isInputPayload()
}

func AsInputPayload[P InputPayload](p InputPayload) (P, bool) {
	pp, ok := p.(P)

	return pp, ok
}
