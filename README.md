# kinesis-tailf

tail -f command for AWS Kinesis.

## Install

```
go get -u github.com/fujiwara/kinesis-tailf/cmd/kinesis-tailf
```

## Usage

Requires `AWS_REGION` environment variable.

```
Usage of kinesis-tail:
  -lf
    	append LF(\n) to each record
  -shard-id string
    	shard id (, separated)
  -stream string
    	stream name
```

kinesis-tailf supports decoding packed records by Kinesis Producer Library.

## Licence

MIT

