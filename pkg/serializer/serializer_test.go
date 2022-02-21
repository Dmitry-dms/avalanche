package serializer

import (
	"testing"
	"time"

	j "github.com/Dmitry-dms/avalanche/pkg/serializer/easyjson"
	"github.com/Dmitry-dms/avalanche/pkg/serializer/json"

)

var (
	stat = j.CompanyStats{
		Name:        "string",
		OnlineUsers: 3,
		MaxUsers:    10,
		TTL:         time.Second,
	}
)

func SerilizeEasyJson(stat j.CompanyStats) ([]byte, error) {
	s := j.CustomEasyJson{}
	return s.Marshal(stat)
}
func SerilizeJson(stat j.CompanyStats) ([]byte, error) {
	s := json.CustomJsonSerializer{}
	return s.Marshal(stat)
}
func BenchmarkEasyJsonSerializeAndDeserialize(b *testing.B) {
	s := j.CustomEasyJson{}
	for i := 0; i < b.N; i++ {
		raw, _ := s.Marshal(stat)
		var d j.CompanyStats
		s.Unmarshal(raw, &d)
		if d.Name != stat.Name {
			b.Fatal("wrong name")
		}
	}
}
func BenchmarkJsonSerializeAndDeserialize(b *testing.B) {
	ser := json.CustomJsonSerializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		raw, _ := ser.Marshal(stat)
		var d j.CompanyStats
		ser.Unmarshal(raw, &d)
		if d.Name != stat.Name {
			b.Fatal("wrong name")
		}
	}
}
func BenchmarkEasyJsonSerialize(b *testing.B) {
	s := j.CustomEasyJson{}
	for i := 0; i < b.N; i++ {
		_, err := s.Marshal(stat)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkJsonSerialize(b *testing.B) {
	ser := json.CustomJsonSerializer{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ser.Marshal(stat)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEasyJsonDeserialize(b *testing.B) {
	s := j.CustomEasyJson{}
	m, _ := SerilizeEasyJson(stat)
	for i := 0; i < b.N; i++ {
		var data j.CompanyStats
		err := s.Unmarshal(m, &data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkJsonDeserialize(b *testing.B) {
	m, _ := SerilizeJson(stat)
	s := json.CustomJsonSerializer{}
	for i := 0; i < b.N; i++ {
		var data j.CompanyStats
		err := s.Unmarshal(m, &data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
