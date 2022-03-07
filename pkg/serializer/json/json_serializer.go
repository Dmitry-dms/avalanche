package json

import "encoding/json"

type CustomJsonSerializer struct {}

func (c *CustomJsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
func (c *CustomJsonSerializer) Unmarshal(msg []byte, v interface{}) error {
	return json.Unmarshal(msg, v)
}
