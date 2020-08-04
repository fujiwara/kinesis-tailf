package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/mashiike/didumean"

	ktail "github.com/fujiwara/kinesis-tailf"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func _main() error {
	var region, streamName, partitionKey string
	var appendLF, firehose bool

	flag.BoolVar(&appendLF, "lf", false, "append LF(\\n) to each record")
	flag.StringVar(&streamName, "stream", "", "stream name")
	flag.StringVar(&partitionKey, "partition-key", "", "partition key")
	flag.StringVar(&region, "region", os.Getenv("AWS_REGION"), "region")
	flag.BoolVar(&firehose, "firehose", false, "put to Firehose delivery stream")
	didumean.Parse()

	if streamName == "" {
		fmt.Fprintln(os.Stderr, "Usage of kinesis-cat:")
		flag.PrintDefaults()
		return nil
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

	fn := app.Cat
	if firehose {
		fn = app.CatFirehose
	}
	if len(flag.Args()) == 0 {
		return fn(ctx, partitionKey, os.Stdin)
	}
	for _, f := range flag.Args() {
		var src io.ReadCloser
		var err error
		if f == "-" {
			src = os.Stdin
		} else if src, err = os.Open(f); err != nil {
			return err
		}
		if err := fn(ctx, partitionKey, src); err != nil {
			return err
		}
		src.Close()
	}
	return nil
}
