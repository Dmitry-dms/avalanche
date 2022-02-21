package easyjson

import (
	"errors"
	"fmt"
	"github.com/mailru/easyjson"
)

type CustomEasyJson struct{}

func (c *CustomEasyJson) Marshal(v interface{}) ([]byte, error) {
	switch t := v.(type) {
	case easyjson.Marshaler:
		return easyjson.Marshal(t)
	default:
		return nil, errors.New(fmt.Sprintf("type - {%s} doesn't implement easyjson.Marshaler interface", t))
	}
}

func (c *CustomEasyJson) Unmarshal(data []byte, v interface{}) error {
	switch t := v.(type) {
	case easyjson.Unmarshaler:
		return easyjson.Unmarshal(data, t)
	default:
		return errors.New(fmt.Sprintf("type - {%s} doesn't implement easyjson.Unmarshaler interface", t))
	}
}