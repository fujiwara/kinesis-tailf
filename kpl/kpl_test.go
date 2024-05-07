package kpl_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/fujiwara/kinesis-tailf/kpl"
)

var ar = &kpl.AggregatedRecord{
	PartitionKeyTable:    []string{"foo", "bar"},
	ExplicitHashKeyTable: []string{"foo", "bar"},
	Records: []*kpl.Record{
		{
			PartitionKeyIndex:    Uint64(0),
			ExplicitHashKeyIndex: Uint64(1),
			Data:                 []byte("data1"),
		},
		{
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
		t.Errorf("roundtrip failed %s %s", string(j1), string(j2))
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

func TestAddDataWithNoPartitionKey(t *testing.T) {
	ar := kpl.NewAggregatedRecord()
	for i := 0; i < 100; i++ {
		ar.AddData([]byte(fmt.Sprintf("data%d", i%10)), "")
	}
	b, err := kpl.Marshal(ar)
	if err != nil {
		t.Error(err)
	}
	restored, err := kpl.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if len(restored.PartitionKeyTable) != 10 {
		t.Errorf("PartitionKeyTable length mismatch got:%d expected:%d", len(restored.PartitionKeyTable), 10)
	}
	for i, rec := range restored.Records {
		if rec.PartitionKeyIndex == nil {
			t.Errorf("PartitionKeyIndex is nil %d", i)
		}
		if !bytes.Equal(rec.Data, []byte(fmt.Sprintf("data%d", i%10))) {
			t.Errorf("Data mismatch restored:%s expected:%s", string(rec.Data), fmt.Sprintf("data%d", i%10))
		}
	}
}

func TestAddDataWithPartitionKey(t *testing.T) {
	ar := kpl.NewAggregatedRecord()
	for i := 0; i < 100; i++ {
		ar.AddData([]byte(fmt.Sprintf("data%d", i%10)), "foo")
	}
	b, err := kpl.Marshal(ar)
	if err != nil {
		t.Error(err)
	}
	restored, err := kpl.Unmarshal(b)
	if err != nil {
		t.Error(err)
	}
	if len(restored.PartitionKeyTable) != 1 {
		t.Errorf("PartitionKeyTable length mismatch got:%d expected:%d", len(restored.PartitionKeyTable), 10)
	}
	for i, rec := range restored.Records {
		if rec.PartitionKeyIndex == nil {
			t.Errorf("PartitionKeyIndex is nil %d", i)
		}
		if !bytes.Equal(rec.Data, []byte(fmt.Sprintf("data%d", i%10))) {
			t.Errorf("Data mismatch restored:%s expected:%s", string(rec.Data), fmt.Sprintf("data%d", i%10))
		}
	}
}
