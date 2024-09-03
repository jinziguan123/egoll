package codec

import (
	"errors"
	"io"
	"syscall"
)

var ErrNotEnough = errors.New("not enough bytes")

// Buffer缓冲区，每个tcp对应一个
type Buffer struct {
	buf   []byte // 应用内缓存区
	start int    // 有效字节开始位置
	end   int    // 有效字节结束位置
}

// New Buffer
func NewBuffer(bytes []byte) *Buffer {
	return &Buffer{
		buf:   bytes,
		start: 0,
		end:   0,
	}
}

// 返回有效字节长度
func (b *Buffer) Len() int {
	return b.end - b.start
}

// 返回容量
func (b *Buffer) Cap() int {
	return len(b.buf)
}

func (b *Buffer) GetBytes() []byte {
	return b.buf[b.start:b.end]
}

func (b *Buffer) GetBuf() []byte {
	return b.buf
}

func (b *Buffer) reset() {
	if b.start == 0 {
		return
	}
	copy(b.buf, b.buf[b.start:b.end])
	b.end -= b.start
	b.start = 0
}

// 从文件描述符中读取数据
func (b *Buffer) ReadFromFD(fd int) error {
	b.reset()

	n, err := syscall.Read(fd, b.buf[b.end:])
	if err != nil {
		return err
	}
	if n == 0 {
		return syscall.EAGAIN
	}
	b.end += n
	return nil
}

// 从reader中读取数据，如果reader阻塞，函数也会阻塞
func (b *Buffer) ReadFromReader(reader io.Reader) (int, error) {
	b.reset()
	n, err := reader.Read(b.buf[b.end:])
	if err != nil {
		return n, err
	}
	b.end += n
	return n, nil
}

// 返回n个字节而不产生位移，如果不够n个字节就返回err
func (b *Buffer) Seek(len int) ([]byte, error) {
	if b.Len() >= len {
		buf := b.buf[b.start : b.start+len]
		return buf, nil
	}
	return nil, ErrNotEnough
}

// 舍弃offset个字段，读取n个字段,如果没有足够的字节，调用reset之后，返回的字节数组失效
func (b *Buffer) Read(offset, limit int) ([]byte, error) {
	if b.Len() < offset+limit {
		return nil, ErrNotEnough
	}
	b.start += offset
	buf := b.buf[b.start : b.start+limit]
	b.start += limit
	return buf, nil
}

// 读取所有字节
func (b *Buffer) ReadAll() []byte {
	buf, _ := b.Read(b.start, b.end)
	return buf
}
