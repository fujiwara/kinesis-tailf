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

type IterateParams struct {
	StreamName     string
	ShardID        string
	StartTimestamp time.Time
	EndTimestamp   time.Time
}

type timeOverFunc func(time.Time) bool

//go:generate protoc --go_out=plugins=kpl:kpl ./kpl.proto

func Iterate(k *kinesis.Kinesis, ch chan []byte, p IterateParams) error {
	in := &kinesis.GetShardIteratorInput{
		ShardId:    aws.String(p.ShardID),
		StreamName: aws.String(p.StreamName),
	}
	if p.StartTimestamp.IsZero() {
		in.ShardIteratorType = aws.String("LATEST")
	} else {
		in.ShardIteratorType = aws.String("AT_TIMESTAMP")
		in.Timestamp = &(p.StartTimestamp)
	}

	var isTimeOver timeOverFunc
	if p.EndTimestamp.IsZero() {
		isTimeOver = func(t time.Time) bool {
			return false
		}
	} else {
		isTimeOver = func(t time.Time) bool {
			return p.EndTimestamp.Before(t)
		}
	}

	r, err := k.GetShardIterator(in)
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
			if isTimeOver(*record.ApproximateArrivalTimestamp) {
				return nil
			}
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
			if isTimeOver(time.Now()) {
				return nil
			}
			time.Sleep(Interval)
		}
	}
}
