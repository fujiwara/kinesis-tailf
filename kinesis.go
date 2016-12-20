package ktail

import (
	"bytes"
	"crypto/md5"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/fujiwara/kinesis-tailf/kpl"
	"github.com/golang/protobuf/proto"
)

var (
	PackedHeader       = []byte{0xF3, 0x89, 0x9A, 0xC2}
	PackedHeaderLength = len(PackedHeader)
	PackedFooterLength = md5.Size
	Interval           = time.Second
	LF                 = []byte{'\n'}
)

//go:generate protoc --go_out=plugins=kpl:kpl ./kpl.proto

func Iterate(k *kinesis.Kinesis, streamName, shardId string, ch chan []byte) error {
	r, err := k.GetShardIterator(&kinesis.GetShardIteratorInput{
		ShardId:           aws.String(shardId),
		ShardIteratorType: aws.String("LATEST"),
		StreamName:        aws.String(streamName),
	})
	if err != nil {
		return err
	}
	itr := r.ShardIterator
	for {
		rr, err := k.GetRecords(&kinesis.GetRecordsInput{
			Limit:         aws.Int64(1000),
			ShardIterator: itr,
		})
		if err != nil {
			return err
		}
		itr = rr.NextShardIterator
		for _, record := range rr.Records {
			rs, err := UnmarshalRecords(record.Data)
			if err == nil {
				for _, r := range rs {
					ch <- r.Data
				}
			} else {
				ch <- record.Data
			}
		}
		if len(rr.Records) == 0 {
			time.Sleep(Interval)
		}
	}
	return nil
}

func UnmarshalRecords(raw []byte) ([]*kpl.Record, error) {
	if bytes.HasPrefix(raw, PackedHeader) {
		var ar kpl.AggregatedRecord
		err := proto.Unmarshal(raw[PackedHeaderLength:len(raw)-PackedFooterLength], &ar)
		if err != nil {
			return nil, err
		}
		return ar.Records, nil
	}
	return nil, errors.New("not_marshaled_data")
}
