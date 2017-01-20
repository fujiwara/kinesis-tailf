package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/apex/go-apex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/fujiwara/kinesis-tailf/kpl"
)

const MaxMessageSize = 64 * 1024
const BufferSize = 4096

type KinesisEvent struct {
	Records []struct {
		Kinesis struct {
			KinesisSchemaVersion        string  `json:"kinesisSchemaVersion"`
			PartitionKey                string  `json:"partitionKey"`
			SequenceNumber              string  `json:"sequenceNumber"`
			Data                        []byte  `json:"data"`
			ApproximateArrivalTimestamp float64 `json:"approximateArrivalTimestamp"`
		} `json:"kinesis"`
		EventSource       string `json:"eventSource"`
		EventVersion      string `json:"eventVersion"`
		EventID           string `json:"eventID"`
		EventName         string `json:"eventName"`
		InvokeIdentityArn string `json:"invokeIdentityArn"`
		AwsRegion         string `json:"awsRegion"`
		EventSourceARN    string `json:"eventSourceARN"`
	} `json:"Records"`
}

var (
	ksv        *kinesis.Kinesis
	streamName *string
	matcher    *regexp.Regexp
)

func init() {
	ksv = kinesis.New(session.New())
	streamName = aws.String(os.Getenv("STREAM_NAME"))
	matcher = regexp.MustCompile(os.Getenv("MATCH"))
	log.Printf("streamName:%s matcher:%s", *streamName, matcher)
}

func main() {
	apex.HandleFunc(proceess)
}

func proceess(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
	var e KinesisEvent
	if err := json.Unmarshal(event, &e); err != nil {
		return nil, err
	}
	if len(e.Records) == 0 {
		return nil, nil
	}

	result := make([]byte, 0, BufferSize)
	count := 0
	match := 0
	for _, record := range e.Records {
		data := record.Kinesis.Data
		ar, err := kpl.Unmarshal(data)
		if err != nil {
			count++
			if matcher.Match(data) {
				match++
				result, err = push(result, data)
				if err != nil {
					return nil, err
				}
			}
			continue
		}
		for _, r := range ar.Records {
			count++
			if matcher.Match(r.Data) {
				match++
				result, err = push(result, r.Data)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	log.Printf("records: event:%d processed:%d match:%d",
		len(e.Records),
		count,
		match,
	)
	if len(result) == 0 {
		return nil, nil
	}
	return nil, flush(result)
}

func push(b, a []byte) ([]byte, error) {
	if len(b)+len(a)+1 >= MaxMessageSize {
		if err := flush(b); err != nil {
			return b, err
		}
		b = make([]byte, 0, BufferSize)
	}
	b = append(b, a...)
	b = append(b, '\n')
	return b, nil
}

func flush(b []byte) error {
	key := fmt.Sprintf("%x", md5.Sum(b))
	log.Printf("putRecord to %s size:%d partitionKey:%s", *streamName, len(b), key)
	_, err := ksv.PutRecord(&kinesis.PutRecordInput{
		Data:         b,
		PartitionKey: aws.String(key),
		StreamName:   streamName,
	})
	return err
}
