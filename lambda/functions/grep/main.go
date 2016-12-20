package main

import (
	"encoding/json"
	"log"
	"os"
	"regexp"

	ktail "github.com/fujiwara/kinesis-tailf"

	"github.com/apex/go-apex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

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
	ch         chan []byte
	ksv        *kinesis.Kinesis
	streamName *string
	matcher    *regexp.Regexp
)

func init() {
	ksv = kinesis.New(session.New())
	streamName = aws.String(os.Getenv("STREAM_NAME"))
	matcher = regexp.MustCompile(os.Getenv("MATCH"))
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

	result := make([]byte, 4096)
	count := 0
	match := 0
	for _, record := range e.Records {
		data := record.Kinesis.Data
		rs, err := ktail.UnmarshalRecords(data)
		if err != nil {
			count++
			if matcher.Match(data) {
				match++
				result = append(result, data...)
				result = append(result, ktail.LF...)
			}
			continue
		}
		for _, r := range rs {
			count++
			if matcher.Match(r.Data) {
				match++
				result = append(result, r.Data...)
				result = append(result, ktail.LF...)
			}
		}
	}
	if len(result) == 0 {
		return nil, nil
	}
	log.Printf("records: event: %d processed: %d match: %d",
		len(e.Records),
		count,
		match,
	)
	return ksv.PutRecord(&kinesis.PutRecordInput{
		Data:         result,
		PartitionKey: aws.String(e.Records[0].Kinesis.PartitionKey),
		StreamName:   streamName,
	})
}
