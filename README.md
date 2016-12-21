# kinesis-tailf

tail -f command for Amazon Kinesis.

## Install

```
go get -u github.com/fujiwara/kinesis-tailf/cmd/kinesis-tailf
```

## Usage

Requires `AWS_REGION` environment variable.

```
Usage of kinesis-tail:
  -lf
    	append LF(\n) to each record data
  -shard-id string
    	shard ids (, separated)
  -stream string
    	stream name
```

kinesis-tailf supports decoding packed records by Kinesis Producer Library.

## Licence

MIT

