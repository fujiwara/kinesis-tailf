package main

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
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
	ksv        *kinesis.Client
	streamName *string
	matcher    *regexp.Regexp
)

func init() {
	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	ksv = kinesis.NewFromConfig(awsCfg)
	streamName = aws.String(os.Getenv("STREAM_NAME"))
	matcher = regexp.MustCompile(os.Getenv("MATCH"))
	log.Printf("streamName:%s matcher:%s", *streamName, matcher)
}

func main() {
	lambda.Start(proceess)
}

func proceess(ctx context.Context, e KinesisEvent) error {
	if len(e.Records) == 0 {
		return nil
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
				result, err = push(ctx, result, data)
				if err != nil {
					return err
				}
			}
			continue
		}
		for _, r := range ar.Records {
			count++
			if matcher.Match(r.Data) {
				match++
				result, err = push(ctx, result, r.Data)
				if err != nil {
					return err
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
		return nil
	}
	return flush(ctx, result)
}

func push(ctx context.Context, b, a []byte) ([]byte, error) {
	if len(b)+len(a)+1 >= MaxMessageSize {
		if err := flush(ctx, b); err != nil {
			return b, err
		}
		b = make([]byte, 0, BufferSize)
	}
	b = append(b, a...)
	b = append(b, '\n')
	return b, nil
}

func flush(ctx context.Context, b []byte) error {
	key := fmt.Sprintf("%x", md5.Sum(b))
	log.Printf("putRecord to %s size:%d partitionKey:%s", *streamName, len(b), key)
	_, err := ksv.PutRecord(ctx, &kinesis.PutRecordInput{
		Data:         b,
		PartitionKey: aws.String(key),
		StreamName:   streamName,
	})
	return err
}
