package ktail

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/fujiwara/kinesis-tailf/kpl"
)

var (
	Interval = time.Second
	LF       = []byte{'\n'}
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
			ar, err := kpl.Unmarshal(record.Data)
			if err == nil {
				for _, r := range ar.Records {
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
}
