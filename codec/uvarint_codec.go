package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
)

type uvarintDecoder struct {
}

func NewUvarintDecoder() Decoder {
	return &uvarintDecoder{}
}

func (d *uvarintDecoder) Decode(buffer *Buffer, handle func([]byte)) error {
	for {
		bytes := buffer.GetBytes()
		bodyLen, headerLen := binary.Uvarint(bytes)
		if int(bodyLen)+headerLen > buffer.Cap() {
			return errors.New(fmt.Sprintf("illegal body length %d", bodyLen))
		}
		body, err := buffer.Read(headerLen, int(bodyLen))
		if errors.Is(err, ErrNotEnough) {
			return nil
		}
		handle(body)
	}
}

type uvarintEncoder struct {
	writeBufferLen  int
	writeBufferPool *sync.Pool
}

// NewUvarintEncoder 创建基于Uvarint的编码器
// writeBufferLen 服务器发送给客户端包的建议长度，当发送的包小于这个值时，会利用到内存池优化
func NewUvarintEncoder(writeBufferLen int) *uvarintEncoder {
	if writeBufferLen <= 0 {
		panic("writeBufferLen must be greater than 0")
	}
	return &uvarintEncoder{
		writeBufferLen: writeBufferLen,
		writeBufferPool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, writeBufferLen)
			},
		},
	}
}

func getUvarintLen(x uint64) int {
	i := 0
	for x >= 0x80 {
		x >>= 7
		i++
	}
	return i + 1
}

// EncodeToWriter 编码数据,并且写入Writer
func (e uvarintEncoder) EncodeToWriter(w io.Writer, bytes []byte) error {
	bytesLen := uint64(len(bytes))
	uvarintLen := getUvarintLen(bytesLen)

	var buffer []byte
	l := uvarintLen + len(bytes)
	if l > e.writeBufferLen {
		buffer = make([]byte, l)
	} else {
		obj := e.writeBufferPool.Get()
		defer e.writeBufferPool.Put(obj)
		buffer = obj.([]byte)[0:l]
	}
	binary.PutUvarint(buffer, bytesLen)
	copy(buffer[uvarintLen:], bytes)
	_, err := w.Write(buffer)
	return err
}
