package core

import "encoding/binary"

type BlockNumber uint64

func (b BlockNumber) ToBytes() []byte {
	numBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(numBytes, uint64(b))

	return numBytes
}
