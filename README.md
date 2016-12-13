# kinesis-tail

tail command for AWS Kinesis.

## Install

```
go get -u github.com/fujiwara/kinesis-tail
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

kinesis-tail supports decoding packed records by Kinesis Producer Library.

## Licence

MIT

