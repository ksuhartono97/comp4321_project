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
func encode64Bit(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

//decode64Bit converts byte slice to unsgined 64 bit
func decode64Bit(b []byte) uint64 {
	return binary.LittleEndian.Uint64(b)
}

//encode32Bit converts unsgined 32 bit to byte slice
func encode32Bit(n uint32) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

//decode32Bit converts byte slice to unsgined 32 bit
func decode32Bit(b []byte) uint32 {
	return binary.LittleEndian.Uint32(b)
}
