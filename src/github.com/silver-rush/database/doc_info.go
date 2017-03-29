package database

import (
	"encoding/binary"
	"fmt"

	"github.com/boltdb/bolt"
)

var docInfoDB *bolt.DB

//OpenDocInfoDB opens the document information database
func OpenDocInfoDB() {
	var err error
	docInfoDB, err = bolt.Open("doc_info.db", 0600, nil)
	if err != nil {
		panic(fmt.Errorf("Open document information databse error: %s", err))
	}

	docInfoDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("doc_info"))
		if err != nil {
			panic(fmt.Errorf("Create doc OpenDocInfoDB bucket error: %s", err))
		}

		return nil
	})
}

//CloseDocInfoDB will close the document information database
func CloseDocInfoDB() {
	docInfoDB.Close()
}

//DocInfo is a struct containing all the document information
type DocInfo struct {
	size     uint32
	time     uint32
	parentID uint64
	childNum uint32
	child    []uint64
	title    string
}

func encodeDocInfo(d *DocInfo) []byte {
	b := make([]byte, 4+4+8+4+8*int(d.childNum)+4*len(d.title))

	binary.LittleEndian.PutUint32(b[0:5], d.size)
	binary.LittleEndian.PutUint32(b[5:9], d.time)
	binary.LittleEndian.PutUint64(b[9:17], d.parentID)
	binary.LittleEndian.PutUint32(b[17:21], d.childNum)
	for i, id := range d.child {
		binary.LittleEndian.PutUint64(b[21+i*8:21+(i+1)*8], id)
	}
	copy(b[21+d.childNum*8:], []byte(d.title))
	return b
}

func decodeDocInfo(b []byte) *DocInfo {
	var d DocInfo

	d.size = binary.LittleEndian.Uint32(b[0:5])
	d.time = binary.LittleEndian.Uint32(b[5:9])
	d.parentID = binary.LittleEndian.Uint64(b[9:17])
	d.childNum = binary.LittleEndian.Uint32(b[17:21])
	var i uint32
	for i = 0; i < d.childNum; i++ {
		d.child = append(d.child, binary.LittleEndian.Uint64(b[21+i*8:21+(i+1)*8]))
	}
	d.title = string(b[21+d.childNum*8:])
	return &d
}

//InsertDocInfo a document info given the document id
func InsertDocInfo(docID uint64, d *DocInfo) {
	docInfoDB.Batch(func(tx *bolt.Tx) error {
		docInfoBucket := tx.Bucket([]byte("doc_info"))

		docInfoBucket.Put(encode64Bit(docID), encodeDocInfo(d))
		return nil
	})
}

//GetDocInfo returns the information relevant to the document id, return nil if not found
func GetDocInfo(docID uint64) *DocInfo {
	var d *DocInfo
	postingDB.View(func(tx *bolt.Tx) error {
		docInfoBucket := tx.Bucket([]byte("doc_info"))

		returnByte := docInfoBucket.Get(encode64Bit(docID))
		if returnByte != nil {
			d = decodeDocInfo(returnByte)
		}
		return nil
	})
	return d
}
