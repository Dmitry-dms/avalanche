package main

import (
	"log"

	"github.com/Dmitry-dms/avalanche/pkg/serializer/json")


func main() {
	var r redisMessage
	m := "{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"
	s := &json.CustomJsonSerializer{}
	err := s.Deserialize([]byte(m), &r)
	if err != nil {
		log.Println(err)
		return // TODO: Handle error
	}
	log.Printf("%s %s %s", r.ClientId, r.CompanyName, r.Message)
}
type redisMessage struct {
	CompanyName string `json:"company_name"`
	ClientId    string `json:"client_id"`
	Message     string `json:"message"`
}//"{\"company_name\":\"testing\",\"client_id\":\"4\",\"message\":\"10\"}"
