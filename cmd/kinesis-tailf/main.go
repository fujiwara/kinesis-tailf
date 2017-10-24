package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/tkuchiki/parsetime"

	ktail "github.com/fujiwara/kinesis-tailf"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func _main() error {
	var region, streamName, shardKey, start, end string
	var startTs, endTs time.Time
	var appendLF bool

	flag.BoolVar(&appendLF, "lf", false, "append LF(\\n) to each record")
	flag.StringVar(&streamName, "stream", "", "stream name")
	flag.StringVar(&shardKey, "shard-key", "", "shard key")
	flag.StringVar(&region, "region", os.Getenv("AWS_REGION"), "region")
	flag.StringVar(&start, "start", "", "start timestamp")
	flag.StringVar(&end, "end", "", "end timestamp")
	flag.Parse()

	if streamName == "" {
		fmt.Fprintln(os.Stderr, "Usage of kinesis-tailf:")
		flag.PrintDefaults()
		return nil
	}

	var err error
	startTs, err = parseTimestamp(start)
	if err != nil {
		return err
	}
	endTs, err = parseTimestamp(end)
	if err != nil {
		return err
	}

	var sess *session.Session
	if region != "" {
		sess = session.New(
			&aws.Config{Region: aws.String(region)},
		)
	} else {
		sess = session.New()
	}

	ctx := context.Background()
	app := ktail.New(sess, streamName)
	app.AppendLF = appendLF
	return app.Run(ctx, shardKey, startTs, endTs)
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
