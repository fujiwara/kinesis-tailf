# kinesis-tailf

tail -f command for Amazon Kinesis.

## Install

### Homebrew

```
$ brew install fujiwara/tap/kinesis-tailf
```

### Binary packages

[Releases](https://github.com/fujiwara/kinesis-tailf/releases).

## Usage

Required flags are below.

- `-stream`
- `-region` or `AWS_REGION` environment variable

```
Usage of kinesis-tailf:
  -end string
    	end timestamp
  -lf
    	append LF(\n) to each record
  -region string
    	region (default AWS_REGION environment variable)
  -shard-key string
    	shard key
  -start string
    	start timestamp
  -stream string
    	stream name
```

kinesis-tailf supports decoding packed records by Kinesis Producer Library (KPL).

## Licence

MIT

