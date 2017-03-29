package database

import "encoding/binary"

//UintToByte converts unsgined 64 bit to byte slice
func UintToByte(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

//ByteToUint converts byte slice to unsgined 64 bit
func ByteToUint(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}
