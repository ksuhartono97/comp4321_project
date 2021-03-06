package database

import (
	"encoding/binary"
	"fmt"
	"os"

	"../../../github.com/boltdb/bolt"
)

var docInfoDB *bolt.DB

//OpenDocInfoDB opens the document information database
func OpenDocInfoDB() {
	var err error
	docInfoDB, err = bolt.Open("db"+string(os.PathSeparator)+"doc_info.db", 0700, nil)
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

//OpenDocInfoDBReadOnly opens the document information database in read-only mode
func OpenDocInfoDBReadOnly() {
	var err error
	docInfoDB, err = bolt.Open("db"+string(os.PathSeparator)+"doc_info.db", 0700, &bolt.Options{ReadOnly: true})
	if err != nil {
		panic(fmt.Errorf("Open document information databse error: %s", err))
	}
}

//CloseDocInfoDB will close the document information database
func CloseDocInfoDB() {
	docInfoDB.Close()
}

//DocInfo is a struct containing all the document information
type DocInfo struct {
	Size      int32
	Time      int64
	ParentNum int32
	Parent    []int64
	ChildNum  int32
	Child     []int64
	Title     string
}

func encodeDocInfo(d *DocInfo) []byte {
	titleByte := []byte(d.Title)
	b := make([]byte, 4+8+4+8*int(d.ParentNum)+4+8*int(d.ChildNum)+len(titleByte))

	binary.LittleEndian.PutUint32(b[0:4], uint32(d.Size))
	binary.LittleEndian.PutUint64(b[4:12], uint64(d.Time))
	binary.LittleEndian.PutUint32(b[12:16], uint32(d.ParentNum))

	for i, id := range d.Parent {
		binary.LittleEndian.PutUint64(b[16+i*8:16+(i+1)*8], uint64(id))
	}
	end := 16 + (len(d.Parent))*8

	binary.LittleEndian.PutUint32(b[end:end+4], uint32(d.ChildNum))
	for i, id := range d.Child {
		binary.LittleEndian.PutUint64(b[end+4+i*8:end+4+(i+1)*8], uint64(id))
	}
	end = end + 4 + (len(d.Child))*8

	copy(b[end:], titleByte)
	return b
}

func decodeDocInfo(b []byte) *DocInfo {
	var d DocInfo

	d.Size = int32(binary.LittleEndian.Uint32(b[0:4]))
	d.Time = int64(binary.LittleEndian.Uint64(b[4:12]))
	d.ParentNum = int32(binary.LittleEndian.Uint32(b[12:16]))
	for i := 0; i < int(d.ParentNum); i++ {
		d.Parent = append(d.Parent, int64(binary.LittleEndian.Uint64(b[16+i*8:16+(i+1)*8])))
	}
	end := 16 + int(d.ParentNum)*8

	d.ChildNum = int32(binary.LittleEndian.Uint32(b[end : end+4]))
	for i := 0; i < int(d.ChildNum); i++ {
		d.Child = append(d.Child, int64(binary.LittleEndian.Uint64(b[end+4+i*8:end+4+(i+1)*8])))
	}

	end = end + 4 + int(d.ChildNum)*8
	d.Title = string(b[end:])
	return &d
}

//InsertDocInfo a document info given the document id
func InsertDocInfo(docID int64, d *DocInfo) {
	err := docInfoDB.Batch(func(tx *bolt.Tx) error {
		docInfoBucket := tx.Bucket([]byte("doc_info"))

		docInfoBucket.Put(encode64Bit(docID), encodeDocInfo(d))
		return nil
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

//GetDocInfo returns the information relevant to the document id, return nil if not found
func GetDocInfo(docID int64) *DocInfo {
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
func GetAllDoc() []int64 {
	var list []int64

	err := docInfoDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("doc_info"))
		list = make([]int64, bucket.Stats().KeyN)
		i := 0

		if err := bucket.ForEach(func(k, v []byte) error {
			list[i] = decode64Bit(k)
			if list[i] != 0 {
				i++
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return list
}
