package database

import (
	"encoding/binary"
	"fmt"
	"os"

	"../../../github.com/boltdb/bolt"
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
	TermFreq  int32
	Positions []int32
}

func encodePosting(p *Posting) []byte {
	b := make([]byte, 4+len(p.Positions)*4)
	binary.LittleEndian.PutUint32(b, uint32(p.TermFreq))
	for i, pos := range p.Positions {
		binary.LittleEndian.PutUint32(b[4+i*4:4+(i+1)*4], uint32(pos))
	}
	return b
}

func decodePosting(b []byte) *Posting {
	var p Posting
	p.TermFreq = int32(binary.LittleEndian.Uint32(b))
	for i := 4; i < len(b); i = i + 4 {
		p.Positions = append(p.Positions, int32(binary.LittleEndian.Uint32(b[i:i+4])))
	}
	return &p
}

//InsertIntoPostingList a record into the posting list of the given word ID, it will also update the forward list
func InsertIntoPostingList(wordID int64, docID int64, p *Posting) {
	err := postingDB.Batch(func(tx *bolt.Tx) error {
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

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

//GetPosting returns a posting from the database with the given document ID and relevant information, return nil if not found
func GetPosting(wordID int64, docID int64) *Posting {
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
func GetTermsInDoc(docID int64) []int64 {
	var list []int64

	postingDB.View(func(tx *bolt.Tx) error {
		allForwardBucket := tx.Bucket([]byte("forward"))
		forwardBucket := allForwardBucket.Bucket(encode64Bit(docID))

		if forwardBucket == nil {
			//Bucket does not exist, return empty list as is
			return nil
		}

		if err := forwardBucket.ForEach(func(k, v []byte) error {
			//Skipping the zeroth index... which is weird for some reason?
			id := decode64Bit(k)
			if id != 0 {
				list = append(list, id)
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	})

	return list
}
