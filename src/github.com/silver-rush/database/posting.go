package database

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var postingDB *bolt.DB

//OpenPostingDB opens the posting list database
func OpenPostingDB() {
	var err error
	postingDB, err = bolt.Open("db"+string(os.PathSeparator)+"posting_list.db", 0600, nil)
	if err != nil {
		panic(fmt.Errorf("Open Posting List Database error: %s", err))
	}

	postingDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("posting"))
		if err != nil {
			panic(fmt.Errorf("Create posting list bucket error: %s", err))
		}

		_, err = tx.CreateBucketIfNotExists([]byte("forward"))
		if err != nil {
			panic(fmt.Errorf("Create forward list bucket error: %s", err))
		}

		return nil
	})
}

//ClosePostingDB will close the posting list database
func ClosePostingDB() {
	postingDB.Close()
}

//Posting is a data struct in the posting list
type Posting struct {
	//docID uint64
	TermFreq uint32
}

//NewPosting make and initialize a posting
func NewPosting() *Posting {
	var p Posting
	p.TermFreq = 0
	return &p
}

func encodePosting(p *Posting) []byte {
	b := make([]byte, 4)
	//binary.LittleEndian.PutUint64(b, p.docID)
	binary.LittleEndian.PutUint32(b, p.TermFreq)
	return b
}

func decodePosting(b []byte) *Posting {
	var p Posting
	//p.docID = binary.LittleEndian.Uint64(b[:9])
	p.TermFreq = binary.LittleEndian.Uint32(b)
	return &p
}

//InsertIntoPostingList a record into the posting list of the given word ID, it will also update the forward list
func InsertIntoPostingList(wordID uint64, docID uint64, p *Posting) {
	postingDB.Batch(func(tx *bolt.Tx) error {
		allPostingBucket := tx.Bucket([]byte("posting"))
		postingBucket, err := allPostingBucket.CreateBucketIfNotExists(encode64Bit(wordID))
		if err != nil {
			fmt.Println("Create Posting bucket error.")
			return err
		}

		allForwardBucket := tx.Bucket([]byte("forward"))
		forwardBucket, err := allForwardBucket.CreateBucketIfNotExists(encode64Bit(docID))
		if err != nil {
			fmt.Println("Create Forward bucket error.")
			return err
		}

		err = postingBucket.Put(encode64Bit(docID), encodePosting(p))
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = forwardBucket.Put(encode64Bit(wordID), []byte{0})
		if err != nil {
			fmt.Println(err)
			return err
		}

		return nil
	})
}

//GetPosting returns a posting from the database with the given document ID and relevant information, return nil if not found
func GetPosting(wordID uint64, docID uint64) *Posting {
	var p *Posting
	postingDB.View(func(tx *bolt.Tx) error {
		allPostingBucket := tx.Bucket([]byte("posting"))
		listBucket := allPostingBucket.Bucket(encode64Bit(wordID))
		if listBucket != nil {
			returnByte := listBucket.Get(encode64Bit(docID))
			if returnByte != nil {
				p = decodePosting(returnByte)
			}
		}
		return nil
	})
	return p
}

//GetTermsInDoc returns a slice of all terms in the document in the forward index
func GetTermsInDoc(docID uint64) []uint64 {
	var list []uint64

	postingDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("forward"))

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
