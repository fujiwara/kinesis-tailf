# kinesis-tailf

tail -f command for Amazon Kinesis.

## Install

```
go get -u github.com/fujiwara/kinesis-tailf/cmd/kinesis-tailf
```

## Usage

Requires `AWS_REGION` environment variable.

```
Usage of kinesis-tailf:
  -lf
    	append LF(\n) to each record
  -region string
    	region
  -shard-id string
    	shard id (, separated)
  -stream string
    	stream name
```

kinesis-tailf supports decoding packed records by Kinesis Producer Library.

## Licence

MIT

