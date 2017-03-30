package database

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var wordDB *bolt.DB

//OpenWordDB opens the word-id database
func OpenWordDB() {
	var err error
	wordDB, err = bolt.Open("db"+string(os.PathSeparator)+"word_id.db", 0600, nil)
	if err != nil {
		panic(fmt.Errorf("Open word ID error: %s", err))
	}

	wordDB.Update(func(tx *bolt.Tx) error {
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

//CloseWordDB close the word-id database
func CloseWordDB() {
	wordDB.Close()
}

//GetIDWithWord returns a unique id for the word. If the record does not exist, it will create one.
func GetIDWithWord(word string) (id int64, created bool) {
	id = 0
	created = false

	err := wordDB.Batch(func(tx *bolt.Tx) error {

		wordToIDBuc := tx.Bucket([]byte("word_to_id"))
		returnByte := wordToIDBuc.Get([]byte(word))

		if returnByte == nil {
			created = true
			idToWordBuc := tx.Bucket([]byte("id_to_word"))
			nextID, _ := idToWordBuc.NextSequence()
			id = int64(nextID)

			err := idToWordBuc.Put(encode64Bit(id), []byte(word))
			if err != nil {
				return err
			}

			err = wordToIDBuc.Put([]byte(word), encode64Bit(id))
			return err
		}

		id = decode64Bit(returnByte)
		return nil
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return
}

//GetWordWithID returns the id given the word, returns empty string if not found
func GetWordWithID(id int64) (s string) {
	var returnByte []byte
	wordDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("id_to_word"))
		returnByte = bucket.Get(encode64Bit(id))
		return nil
	})

	if returnByte == nil {
		return ""
	}

	return string(returnByte)
}
