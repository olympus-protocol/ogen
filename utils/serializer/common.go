package serializer

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"io"
	"time"
)

type binaryFreeList chan []byte

type Uint32Time time.Time

const (
	CommandSize            = 12
	binaryFreeListMaxItems = 1024
)

var (
	binarySerializer binaryFreeList = make(chan []byte, binaryFreeListMaxItems)
	bigEndian                       = binary.BigEndian
	littleEndian                    = binary.LittleEndian
)

func randomUint64(r io.Reader) (uint64, error) {
	rv, err := binarySerializer.Uint64(r, bigEndian)
	if err != nil {
		return 0, err
	}
	return rv, nil
}

func RandomUint64() (uint64, error) {
	return randomUint64(rand.Reader)
}

func (l binaryFreeList) Borrow() []byte {
	var buf []byte
	select {
	case buf = <-l:
	default:
		buf = make([]byte, 8)
	}
	return buf[:8]
}

func (l binaryFreeList) Return(buf []byte) {
	select {
	case l <- buf:
	default:

	}
}

func (l binaryFreeList) Uint8(r io.Reader) (uint8, error) {
	buf := l.Borrow()[:1]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := buf[0]
	l.Return(buf)
	return rv, nil
}

func (l binaryFreeList) Uint16(r io.Reader, byteOrder binary.ByteOrder) (uint16, error) {
	buf := l.Borrow()[:2]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint16(buf)
	l.Return(buf)
	return rv, nil
}

func (l binaryFreeList) Uint32(r io.Reader, byteOrder binary.ByteOrder) (uint32, error) {
	buf := l.Borrow()[:4]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint32(buf)
	l.Return(buf)
	return rv, nil
}

func (l binaryFreeList) Uint64(r io.Reader, byteOrder binary.ByteOrder) (uint64, error) {
	buf := l.Borrow()[:8]
	if _, err := io.ReadFull(r, buf); err != nil {
		l.Return(buf)
		return 0, err
	}
	rv := byteOrder.Uint64(buf)
	l.Return(buf)
	return rv, nil
}

func (l binaryFreeList) PutUint8(w io.Writer, val uint8) error {
	buf := l.Borrow()[:1]
	buf[0] = val
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

func (l binaryFreeList) PutUint16(w io.Writer, byteOrder binary.ByteOrder, val uint16) error {
	buf := l.Borrow()[:2]
	byteOrder.PutUint16(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

func (l binaryFreeList) PutUint32(w io.Writer, byteOrder binary.ByteOrder, val uint32) error {
	buf := l.Borrow()[:4]
	byteOrder.PutUint32(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

func (l binaryFreeList) PutUint64(w io.Writer, byteOrder binary.ByteOrder, val uint64) error {
	buf := l.Borrow()[:8]
	byteOrder.PutUint64(buf, val)
	_, err := w.Write(buf)
	l.Return(buf)
	return err
}

type Serializable interface {
	Serialize(w io.Writer) error
}

func Hash(d Serializable) chainhash.Hash {
	h := []byte{}
	b := bytes.NewBuffer(h)
	_ = d.Serialize(b)
	var hash chainhash.Hash
	copy(hash[:], h)
	return hash
}
