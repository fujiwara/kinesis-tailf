package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/tkuchiki/parsetime"

	ktail "github.com/fujiwara/kinesis-tailf"
)

var (
	streamName    = ""
	appendLF      = false
	region        = ""
	flushInterval = 100 * time.Millisecond
)

func main() {
	code := _main()
	os.Exit(code)
}

func _main() int {
	shardIds := make([]string, 0)
	var shardId, start, end string
	var startTs, endTs time.Time

	flag.BoolVar(&appendLF, "lf", false, "append LF(\\n) to each record")
	flag.StringVar(&streamName, "stream", "", "stream name")
	flag.StringVar(&shardId, "shard-id", "", "shard id (, separated) default: all shards in the stream.")
	flag.StringVar(&region, "region", os.Getenv("AWS_REGION"), "region")
	flag.StringVar(&start, "start", "", "start timestamp")
	flag.StringVar(&end, "end", "", "end timestamp")
	flag.Parse()

	if streamName == "" {
		fmt.Fprintln(os.Stderr, "Usage of kinesis-tailf:")
		flag.PrintDefaults()
		return 0
	}

	var err error
	startTs, err = parseTimestamp(start)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	endTs, err = parseTimestamp(end)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	var sess *session.Session
	if region != "" {
		sess = session.New(
			&aws.Config{Region: aws.String(region)},
		)
	} else {
		sess = session.New()
	}

	k := kinesis.New(sess)
	sd, err := k.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
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

	var wg, wgW sync.WaitGroup
	ch := make(chan []byte, 100)
	ctx, cancel := context.WithCancel(context.Background())
	wgW.Add(1)
	go writer(ctx, ch, &wgW)

	for _, id := range shardIds {
		wg.Add(1)
		go func(id string) {
			param := ktail.IterateParams{
				StreamName:     streamName,
				ShardID:        id,
				StartTimestamp: startTs,
				EndTimestamp:   endTs,
			}
			err := ktail.Iterate(k, ch, param)
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}(id)
	}
	wg.Wait()
	cancel()
	wgW.Wait()
	return 0
}

func writer(ctx context.Context, ch chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	var mu sync.Mutex

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	// run periodical flusher
	go func() {
		c := time.Tick(flushInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-c:
				mu.Lock()
				w.Flush()
				mu.Unlock()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case b := <-ch:
			mu.Lock()
			w.Write(b)
			if appendLF {
				w.Write(ktail.LF)
			}
			mu.Unlock()
		}
	}
}

func parseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	p, _ := parsetime.NewParseTime()
	if ts, err := p.Parse(s); err != nil {
		return time.Time{}, fmt.Errorf("can't parse timestamp %s %s\n", s, err)
	} else {
		return ts, nil
	}
}
