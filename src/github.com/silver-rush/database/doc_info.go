package database

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var docInfoDB *bolt.DB

//OpenDocInfoDB opens the document information database
func OpenDocInfoDB() {
	var err error
	docInfoDB, err = bolt.Open("db"+string(os.PathSeparator)+"doc_info.db", 0600, nil)
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
	Size     uint32
	Time     uint32
	ParentID uint64
	ChildNum uint32
	Child    []uint64
	Title    string
}

func encodeDocInfo(d *DocInfo) []byte {
	b := make([]byte, 4+4+8+4+8*int(d.ChildNum)+4*len(d.Title))

	binary.LittleEndian.PutUint32(b[0:4], d.Size)
	binary.LittleEndian.PutUint32(b[4:8], d.Time)
	binary.LittleEndian.PutUint64(b[8:16], d.ParentID)
	binary.LittleEndian.PutUint32(b[16:20], d.ChildNum)
	for i, id := range d.Child {
		binary.LittleEndian.PutUint64(b[20+i*8:20+(i+1)*8], id)
	}
	copy(b[20+d.ChildNum*8:], []byte(d.Title))
	return b
}

func decodeDocInfo(b []byte) *DocInfo {
	var d DocInfo

	d.Size = binary.LittleEndian.Uint32(b[0:4])
	d.Time = binary.LittleEndian.Uint32(b[4:8])
	d.ParentID = binary.LittleEndian.Uint64(b[8:16])
	d.ChildNum = binary.LittleEndian.Uint32(b[16:20])
	var i uint32
	for i = 0; i < d.ChildNum; i++ {
		d.Child = append(d.Child, binary.LittleEndian.Uint64(b[20+i*8:20+(i+1)*8]))
	}
	d.Title = string(b[20+d.ChildNum*8:])
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
	docInfoDB.View(func(tx *bolt.Tx) error {
		docInfoBucket := tx.Bucket([]byte("doc_info"))

		returnByte := docInfoBucket.Get(encode64Bit(docID))
		if returnByte != nil {
			d = decodeDocInfo(returnByte)
		}
		return nil
	})
	return d
}

//GetAllDoc returns the slice of all document id
func GetAllDoc() []uint64 {
	var list []uint64

	docInfoDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("doc_info"))
		list := make([]uint64, bucket.Stats().KeyN)
		i := 0

		if err := bucket.ForEach(func(k, v []byte) error {
			list[i] = decode64Bit(k)
			i++
			return nil
		}); err != nil {
			return err
		}
		return nil
	})

	return list
}
