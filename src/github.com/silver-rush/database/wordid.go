package database

import (
	"fmt"

	"github.com/boltdb/bolt"
)

var db *bolt.DB

//OpenWordDB opens the word-id databse
func OpenWordDB() {
	var err error
	db, err = bolt.Open("word_id.db", 0600, nil)
	if err != nil {
		panic(fmt.Errorf("Open word ID error: %s", err))
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("word_to_id"))
		if err != nil {
			panic(fmt.Errorf("Create word to id bucket error: %s", err))
		}

		_, err = tx.CreateBucketIfNotExists([]byte("id_to_word"))
		if err != nil {
			panic(fmt.Errorf("Create id to word bucket error: %s", err))
		}

		return nil
	})
}

//GetWordID returns a unique id for the word. If the record does not exist, it will create one.
func GetWordID(word string) (id uint64, created bool) {
	id = 0
	created = true

	var returnByte []byte
	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("word_to_id"))
		returnByte = bucket.Get([]byte(word))
		return nil
	})

	if returnByte == nil {
		created = true

		db.Batch(func(tx *bolt.Tx) error {
			idToWordBuc := tx.Bucket([]byte("id_to_word"))
			id, _ = idToWordBuc.NextSequence()

			err := idToWordBuc.Put(UintToByte(id), []byte(word))
			if err != nil {
				return err
			}

			wordToIDBuc := tx.Bucket([]byte("word_to_id"))
			err = wordToIDBuc.Put([]byte(word), UintToByte(id))

			return err
		})
	} else {
		id = ByteToUint(returnByte)
	}

	return
}
