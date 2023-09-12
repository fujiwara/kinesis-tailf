package ktail

import (
	"bufio"
	"context"
	"crypto/md5"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"

	"github.com/fujiwara/kinesis-tailf/kpl"
)

var (
	flushInterval   = 100 * time.Millisecond
	iterateInterval = time.Second
	LF              = []byte{'\n'}
)

type IterateParams struct {
	StreamName     string
	ShardID        string
	StartTimestamp time.Time
	EndTimestamp   time.Time
}

type timeOverFunc func(time.Time) bool

//go:generate protoc --go_out=plugins=kpl:kpl ./kpl.proto

type App struct {
	kinesis    *kinesis.Client
	StreamName string
	AppendLF   bool
}

func New(cfg aws.Config, name string) *App {
	return &App{
		kinesis:    kinesis.NewFromConfig(cfg),
		StreamName: name,
	}
}

func (app *App) Run(ctx context.Context, shardKey string, startTs, endTs time.Time) error {
	shardIds, err := app.determinShardIds(ctx, shardKey)
	if err != nil {
		return err
	}

	var wg, wgW sync.WaitGroup
	ch := make(chan []byte, 1000)
	ctxC, cancel := context.WithCancel(ctx)
	wgW.Add(1)
	go app.writer(ctxC, ch, &wgW)

	for _, id := range shardIds {
		wg.Add(1)
		go func(id string) {
			param := IterateParams{
				ShardID:        id,
				StartTimestamp: startTs,
				EndTimestamp:   endTs,
			}
			err := app.iterate(ctx, param, ch)
			if err != nil {
				log.Println(err)
			}
			wg.Done()
		}(id)
	}
	wg.Wait()
	cancel()
	close(ch)
	wgW.Wait()
	return nil
}

func (app *App) iterate(ctx context.Context, p IterateParams, ch chan []byte) error {
	in := &kinesis.GetShardIteratorInput{
		ShardId:    aws.String(p.ShardID),
		StreamName: aws.String(app.StreamName),
	}
	if p.StartTimestamp.IsZero() {
		in.ShardIteratorType = types.ShardIteratorTypeLatest
	} else {
		in.ShardIteratorType = types.ShardIteratorTypeAtTimestamp
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

	r, err := app.kinesis.GetShardIterator(ctx, in)
	if err != nil {
		return err
	}
	itr := r.ShardIterator
	for {
		rr, err := app.kinesis.GetRecords(ctx, &kinesis.GetRecordsInput{
			Limit:         aws.Int32(1000),
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
			time.Sleep(iterateInterval)
		}
	}
}

func (app *App) writer(ctx context.Context, ch chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	var mu sync.Mutex

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	// run periodical flusher
	go func() {
		c := time.Tick(flushInterval)
		for {
			select {
			case <-c:
				mu.Lock()
				w.Flush()
				mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		b, ok := <-ch
		if !ok {
			// channel closed
			return
		}
		mu.Lock()
		w.Write(b)
		if app.AppendLF {
			w.Write(LF)
		}
		mu.Unlock()
	}
}

func toHashKey(s string) *big.Int {
	b := md5.Sum([]byte(s))
	return big.NewInt(0).SetBytes(b[:])
}

func (app *App) determinShardIds(ctx context.Context, shardKey string) ([]string, error) {
	var shardIds []string

	sd, err := app.kinesis.DescribeStream(ctx, &kinesis.DescribeStreamInput{
		StreamName: aws.String(app.StreamName),
	})
	if err != nil {
		return shardIds, err
	}

	if shardKey == "" {
		// all shards
		for _, s := range sd.StreamDescription.Shards {
			shardIds = append(shardIds, *s.ShardId)
		}
		return shardIds, nil
	}

	hashKey := toHashKey(shardKey)

	for _, s := range sd.StreamDescription.Shards {
		start, end := big.NewInt(0), big.NewInt(0)
		start.SetString(*s.HashKeyRange.StartingHashKey, 10)
		end.SetString(*s.HashKeyRange.EndingHashKey, 10)

		if start.Cmp(hashKey) <= 0 && hashKey.Cmp(end) <= 0 {
			shardIds = append(shardIds, *s.ShardId)
			break
		}
	}
	return shardIds, nil
}
