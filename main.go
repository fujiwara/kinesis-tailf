package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"flag"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/fujiwara/kinesis-tail/kpl"
	"github.com/golang/protobuf/proto"
)

var (
	mu              sync.Mutex
	streamName      = ""
	interval        = time.Second
	maxFetchRecords = 1000
	appendLF        = false
	magicNumber     = []byte{0xF3, 0x89, 0x9A, 0xC2}
)

//go:generate protoc --go_out=plugins=kpl:kpl ./kpl.proto

func main() {
	shardIds := make([]string, 0)
	var shardId string
	flag.BoolVar(&appendLF, "lf", false, "append LF(\\n) to each record")
	flag.StringVar(&streamName, "stream", "", "stream name")
	flag.StringVar(&shardId, "shard-id", "", "shard id (, separated)")
	flag.Parse()

	k := kinesis.New(session.New())
	sd, err := k.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		log.Fatal(err)
	}

	if shardId != "" {
		for _, s := range strings.Split(shardId, ",") {
			shardIds = append(shardIds, s)
		}
	} else {
		for _, s := range sd.StreamDescription.Shards {
			shardIds = append(shardIds, *s.ShardId)
		}
	}

	var wg sync.WaitGroup
	ch := make(chan []byte, 100)
	for _, id := range shardIds {
		wg.Add(1)
		go func(id string) {
			err := iterate(k, id, ch)
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}(id)
	}
	wg.Add(1)
	go func() {
		writer(ch)
		wg.Done()
	}()
	wg.Wait()
}

func writer(ch chan []byte) {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for {
		b := <-ch
		w.Write(b)
		if appendLF {
			w.Write([]byte{'\n'})
		}
		w.Flush()
	}
}

func iterate(k *kinesis.Kinesis, shardId string, ch chan []byte) error {
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
			if bytes.HasPrefix(record.Data, magicNumber) {
				rs, err := unmashal(record.Data[len(magicNumber) : len(record.Data)-md5.Size])
				if err != nil {
					log.Println(err)
					continue
				}
				for _, r := range rs {
					ch <- r.Data
				}
			} else {
				ch <- record.Data
			}
		}
		if len(rr.Records) == 0 {
			time.Sleep(interval)
		}
	}
	return nil
}

func unmashal(raw []byte) ([]*kpl.Record, error) {
	var ar kpl.AggregatedRecord
	err := proto.Unmarshal(raw, &ar)
	if err != nil {
		return nil, err
	}
	return ar.Records, nil
}
