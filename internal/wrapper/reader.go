package wrapper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

func NewBodyReader(r io.Reader) (*BodyReader, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("draining reader: %w", err)
	}
	br := &BodyReader{b: bytes.NewReader(buf)}
	return br, nil
}

func NewBodyReaderFromRaw(data []byte) *BodyReader {
	return &BodyReader{b: bytes.NewReader(data)}
}

type BodyReader struct {
	b *bytes.Reader
}

func (r *BodyReader) Read(b []byte) (int, error) {
	n, err := r.b.Read(b)
	if err != nil && err != io.EOF {
		return 0, fmt.Errorf("reading buffer: %w", err)
	}
	return n, err
}

func (r *BodyReader) Close() error {
	if _, err := r.b.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking: %w", err)
	}
	return nil
}
