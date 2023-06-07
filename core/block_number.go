package core

import "encoding/binary"

type BlockNumber uint64

func (b BlockNumber) ToBytes() []byte {
	const lenOfByteSlice = 8

	numBytes := make([]byte, lenOfByteSlice)
	binary.BigEndian.PutUint64(numBytes, uint64(b))

	return numBytes
}
