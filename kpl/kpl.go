package kpl

import (
	"bytes"
	"crypto/md5"
	"errors"

	"google.golang.org/protobuf/proto"
)

var (
	MagicNumber = []byte{0xF3, 0x89, 0x9A, 0xC2}
)

func Unmarshal(b []byte) (*AggregatedRecord, error) {
	if bytes.HasPrefix(b, MagicNumber) {
		var ar AggregatedRecord
		end := len(b) - md5.Size
		msg := b[len(MagicNumber):end]
		checksum := b[end:]
		err := proto.Unmarshal(msg, &ar)
		if err != nil {
			return nil, err
		}
		h := md5.New()
		h.Write(msg)
		if !bytes.Equal(checksum, h.Sum(nil)) {
			return nil, errors.New("checksum mismatch")
		}
		return &ar, nil
	}
	return nil, errors.New("not a marshaled data")
}

func Marshal(ar *AggregatedRecord) ([]byte, error) {
	var b []byte
	b = append(b, MagicNumber...)
	packed, err := proto.Marshal(ar)
	if err != nil {
		return nil, err
	}
	b = append(b, packed...)
	h := md5.New()
	h.Write(packed)
	b = append(b, h.Sum(nil)...)
	return b, nil
}
