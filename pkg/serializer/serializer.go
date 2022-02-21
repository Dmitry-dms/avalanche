package serializer


type Serializer interface {
	Marshal(v interface{}) ([]byte,error)
	Unmarshal(msg []byte, v interface{}) error
}