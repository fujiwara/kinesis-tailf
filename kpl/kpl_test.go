package kpl_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/fujiwara/kinesis-tailf/kpl"
)

func TestRoundTrip(t *testing.T) {
	ar := &kpl.AggregatedRecord{
		PartitionKeyTable:    []string{"foo", "bar"},
		ExplicitHashKeyTable: []string{"foo", "bar"},
		Records: []*kpl.Record{
			&kpl.Record{
				PartitionKeyIndex:    Uint64(1),
				ExplicitHashKeyIndex: Uint64(2),
				Data:                 []byte("data1"),
			},
			&kpl.Record{
				PartitionKeyIndex:    Uint64(3),
				ExplicitHashKeyIndex: Uint64(4),
				Data:                 []byte("data2"),
			},
		},
	}
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

func Uint64(v uint64) *uint64 {
	return &v
}
