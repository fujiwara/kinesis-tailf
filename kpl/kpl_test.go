package kpl_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/fujiwara/kinesis-tailf/kpl"
)

var ar = &kpl.AggregatedRecord{
	PartitionKeyTable:    []string{"foo", "bar"},
	ExplicitHashKeyTable: []string{"foo", "bar"},
	Records: []*kpl.Record{
		&kpl.Record{
			PartitionKeyIndex:    Uint64(0),
			ExplicitHashKeyIndex: Uint64(1),
			Data:                 []byte("data1"),
		},
		&kpl.Record{
			PartitionKeyIndex:    Uint64(0),
			ExplicitHashKeyIndex: Uint64(1),
			Data:                 []byte("data2"),
		},
	},
}
var serialized, _ = kpl.Marshal(ar)

func TestRoundTrip(t *testing.T) {
	b, err := kpl.Marshal(ar)
	if err != nil {
		t.Error(err)
	}
	ar2, err := kpl.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	j1, _ := json.Marshal(ar)
	j2, _ := json.Marshal(ar2)
	if !bytes.Equal(j1, j2) {
		t.Error("roudtrip failed %s %s", string(j1), string(j2))
	}
}

func BenchmarkMarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		kpl.Marshal(ar)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		kpl.Unmarshal(serialized)
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json.Marshal(ar)
	}
}

func Uint64(v uint64) *uint64 {
	return &v
}
