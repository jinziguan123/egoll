package codec

import "io"

type Decoder interface {
	Decode(*Buffer, func([]byte)) error
}

type Encoder interface {
	EncodeToWriter(w io.Writer, bytes []byte) error
}
