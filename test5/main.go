package main

import (
	"fmt"

	"github.com/Dmitry-dms/avalanche/pkg/serializer/json")


func main() {
	m := &redisMessage{
		CompanyName: "test",
		ClientId:    "test",
		Message:     "hello",
	}
	ser := &json.CustomJsonSerializer{}
	b, err := ser.Serialize(m)
	if err != nil {
		fmt.Println(err)
		return // TODO: Handle error
	}
	fmt.Printf("serialized msg: %s",b)
	g := string(b)
	var s redisMessage
	err = ser.Deserialize([]byte(g), &s)
	if err != nil {
		fmt.Println(err)
		return // TODO: Handle error
	}
	fmt.Printf("deserialized msg: %s",s)
}

type redisMessage struct {
	CompanyName string
	ClientId    string
	Message     string
}