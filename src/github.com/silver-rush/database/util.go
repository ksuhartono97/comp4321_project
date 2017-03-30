package database

import "encoding/binary"
import "os"

//OpenAllDatabase will make a connection to open all databases
func OpenAllDatabase() {
	os.Mkdir("db", os.ModeDir)
	OpenURLDB()
	OpenWordDB()
	OpenPostingDB()
	OpenDocInfoDB()
}

//CloseAllDatabase will close all database connection
func CloseAllDatabase() {
	CloseURLDB()
	CloseWordDB()
	ClosePostingDB()
	CloseDocInfoDB()
}

//encode64Bit converts unsgined 64 bit to byte slice
func encode64Bit(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}

//decode64Bit converts byte slice to unsgined 64 bit
func decode64Bit(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}

//encode32Bit converts unsgined 32 bit to byte slice
func encode32Bit(n int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(n))
	return b
}

//decode32Bit converts byte slice to unsgined 32 bit
func decode32Bit(b []byte) int32 {
	return int32(binary.LittleEndian.Uint32(b))
}
