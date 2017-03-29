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
	urlDB, err = bolt.Open("db"+string(os.PathSeparator)+"url_id.db", 0600, nil)
	if err != nil {
		panic(fmt.Errorf("Open URL ID error: %s", err))
	}

	urlDB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("url_to_id"))
		if err != nil {
			panic(fmt.Errorf("Create URL to id bucket error: %s", err))
		}

		_, err = tx.CreateBucketIfNotExists([]byte("id_to_url"))
		if err != nil {
			panic(fmt.Errorf("Create id to URL bucket error: %s", err))
		}

		return nil
	})
}

//CloseURLDB close the url-id database
func CloseURLDB() {
	urlDB.Close()
}

//GetURLID returns a unique id for the URL. If the record does not exist, it will create one.
func GetURLID(url string) (id uint64, created bool) {
	id = 0
	created = true

	var returnByte []byte
	urlDB.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("url_to_id"))
		returnByte = bucket.Get([]byte(url))
		return nil
	})

	if returnByte == nil {
		created = true

		urlDB.Batch(func(tx *bolt.Tx) error {
			idToURLBuc := tx.Bucket([]byte("id_to_url"))
			id, _ = idToURLBuc.NextSequence()

			err := idToURLBuc.Put(encode64Bit(id), []byte(url))
			if err != nil {
				return err
			}

			urlToIDBuc := tx.Bucket([]byte("url_to_id"))
			err = urlToIDBuc.Put([]byte(url), encode64Bit(id))

			return err
		})
	} else {
		id = decode64Bit(returnByte)
	}

	return
}

//GetURLWithID returns the id given the URL, returns empty string if not found
func GetURLWithID(id uint64) (s string) {
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
