package ktail

import (
	"bufio"
	"context"
	"crypto/sha256"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/firehose/types"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
)

func (app *App) Cat(ctx context.Context, partitionKey string, src io.Reader) error {
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		b := scanner.Bytes()
		if app.AppendLF {
			b = append(b, LF...)
		}
		in := &kinesis.PutRecordInput{
			Data:       b,
			StreamName: &app.StreamName,
		}
		if partitionKey == "" {
			pk := fmt.Sprintf("%x", sha256.Sum256(b))
			in.PartitionKey = &pk
		} else {
			in.PartitionKey = &partitionKey
		}
		if _, err := app.kinesis.PutRecord(ctx, in); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (app *App) CatFirehose(ctx context.Context, _ string, src io.Reader) error {
	fh := firehose.NewFromConfig(app.cfg)
	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		b := scanner.Bytes()
		if app.AppendLF {
			b = append(b, LF...)
		}
		in := &firehose.PutRecordInput{
			Record: &types.Record{
				Data: b,
			},
			DeliveryStreamName: &app.StreamName,
		}
		if _, err := fh.PutRecord(ctx, in); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
