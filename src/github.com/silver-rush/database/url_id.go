package database

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var urlDB *bolt.DB

//OpenURLDB opens the url-id database
func OpenURLDB() {
	var err error
	urlDB, err = bolt.Open("db"+string(os.PathSeparator)+"url_id.db", 0700, nil)
	if err != nil {
		panic(fmt.Errorf("Open URL ID error: %s", err))
	}

	urlDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("url_to_id"))
		if err != nil {
			panic(fmt.Errorf("Create URL to id bucket error: %s", err))
		}

		idToURLBuc, err := tx.CreateBucketIfNotExists([]byte("id_to_url"))
		if err != nil {
			panic(fmt.Errorf("Create id to URL bucket error: %s", err))
		}

		totalDocByte := idToURLBuc.Get(encode64Bit(0))
		if totalDocByte == nil {
			//Init the total documents count if not done so
			idToURLBuc.Put(encode64Bit(0), encode32Bit(0))
		}

		return nil
	})
}

//CloseURLDB close the url-id database
func CloseURLDB() {
	urlDB.Close()
}

//GetURLID returns a unique id for the URL. If the record does not exist, it will create one.
func GetURLID(url string) (id int64, created bool) {
	id = 0
	created = false

	urlDB.View(func(tx *bolt.Tx) error {
		urlToIDBuc := tx.Bucket([]byte("url_to_id"))
		returnByte := urlToIDBuc.Get([]byte(url))

		if returnByte == nil {
			//id not found. Need to create one.
			created = true
		} else {
			//ID found.
			id = decode64Bit(returnByte)
		}
		return nil
	})

	if created {
		urlDB.Batch(func(tx *bolt.Tx) error {
			urlToIDBuc := tx.Bucket([]byte("url_to_id"))
			idToURLBuc := tx.Bucket([]byte("id_to_url"))
			nextID, _ := idToURLBuc.NextSequence()
			id = int64(nextID)

			//Add one to the total document count (stored at 0th in idToURLBuc)
			err := idToURLBuc.Put(encode64Bit(0), encode32Bit(decode32Bit(idToURLBuc.Get(encode64Bit(0)))+1))
			if err != nil {
				return err
			}

			err = idToURLBuc.Put(encode64Bit(id), []byte(url))
			if err != nil {
				return err
			}

			err = urlToIDBuc.Put([]byte(url), encode64Bit(id))
			return err
		})
	}

	return
}

//GetURLWithID returns the id given the URL, returns empty string if not found
func GetURLWithID(id int64) (s string) {
	var returnByte []byte
	urlDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("id_to_url"))
		returnByte = bucket.Get(encode64Bit(id))
		return nil
	})

	if returnByte == nil {
		return ""
	}

	return string(returnByte)
}

//GetTotalNumberOfDocument returns the total number of documents in the database
func GetTotalNumberOfDocument() int32 {
	var returnByte []byte
	urlDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("id_to_url"))
		returnByte = bucket.Get(encode64Bit(0))
		return nil
	})

	if returnByte == nil {
		return 0
	}

	return decode32Bit(returnByte)

}
