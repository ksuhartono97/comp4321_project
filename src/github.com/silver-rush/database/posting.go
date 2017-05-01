package database

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"

	"math"

	"../../../github.com/boltdb/bolt"
)

var postingDB *bolt.DB

//OpenPostingDB opens the posting list database
func OpenPostingDB() {
	var err error
	postingDB, err = bolt.Open("db"+string(os.PathSeparator)+"posting_list.db", 0700, nil)
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

		_, err = tx.CreateBucketIfNotExists([]byte("stem_posting"))
		if err != nil {
			panic(fmt.Errorf("Create stemmed posting list bucket error: %s", err))
		}

		_, err = tx.CreateBucketIfNotExists([]byte("stem_forward"))
		if err != nil {
			panic(fmt.Errorf("Create stemmed forward list bucket error: %s", err))
		}

		return nil
	})
}

//OpenPostingDBReadOnly opens the posting list database in read-only mode
func OpenPostingDBReadOnly() {
	var err error
	postingDB, err = bolt.Open("db"+string(os.PathSeparator)+"posting_list.db", 0700, &bolt.Options{ReadOnly: true})
	if err != nil {
		panic(fmt.Errorf("Open Posting List Database error: %s", err))
	}
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

//BatchInsertIntoPostingList insert records in batch
func BatchInsertIntoPostingList(docID int64, records map[int64]*Posting) {
	err := postingDB.Batch(func(tx *bolt.Tx) error {
		allPostingBucket := tx.Bucket([]byte("posting"))

		allForwardBucket := tx.Bucket([]byte("forward"))
		forwardBucket, err := allForwardBucket.CreateBucketIfNotExists(encode64Bit(docID))
		if err != nil {
			fmt.Println("Create Forward bucket error.")
			return err
		}

		var maxTF int32
		maxTF = 0
		for wordID, posting := range records {
			postingBucket, err := allPostingBucket.CreateBucketIfNotExists(encode64Bit(wordID))
			if err != nil {
				return err
			}

			docAlreadyExist := postingBucket.Get(encode64Bit(docID))
			if docAlreadyExist == nil {
				//If the document is not in the posting list before, increase DF by 1
				totalDFByte := postingBucket.Get(encode64Bit(0))
				var totalDF int32
				if totalDFByte != nil {
					//totalTerms stores the current total document frequency
					totalDF = decode32Bit(totalDFByte)
				} else {
					totalDF = 0
				}

				//Insert the total DF back
				err = postingBucket.Put(encode64Bit(0), encode32Bit(totalDF+1))
				if err != nil {
					return err
				}
			}

			err = postingBucket.Put(encode64Bit(docID), encodePosting(posting))
			if err != nil {
				return err
			}

			termAlreadyExist := forwardBucket.Get(encode64Bit(wordID))
			if termAlreadyExist == nil {
				//Increase totalTerms by 1 if the term was not indexed for this document before
				totalTermsByte := forwardBucket.Get(encode64Bit(0))
				var totalTerms int32
				if totalTermsByte != nil {
					//totalTF stores the current total number of terms in the document
					totalTerms = decode32Bit(totalTermsByte)
				} else {
					totalTerms = 0
				}

				//Insert the total DF back
				err = forwardBucket.Put(encode64Bit(0), encode32Bit(totalTerms+1))

				if err != nil {
					return err
				}
			}

			if posting.TermFreq > maxTF {
				maxTF = posting.TermFreq
			}

			err = forwardBucket.Put(encode64Bit(wordID), encode32Bit(posting.TermFreq))
			if err != nil {
				return err
			}
		}

		//ID 0 stores the precomputed max TF
		err = forwardBucket.Put(encode64Bit(0), encode32Bit(maxTF))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

//BatchInsertIntoStemmedList insert a batch of words into the stemmed queue
func BatchInsertIntoStemmedList(docID int64, records map[int64]int32) {
	err := postingDB.Batch(func(tx *bolt.Tx) error {
		allPostingBucket := tx.Bucket([]byte("stem_posting"))

		allForwardBucket := tx.Bucket([]byte("stem_forward"))
		forwardBucket, err := allForwardBucket.CreateBucketIfNotExists(encode64Bit(docID))
		if err != nil {
			fmt.Println("Create Forward bucket error.")
			return err
		}

		var maxTF int32
		maxTF = 0
		for wordID, termFreq := range records {
			postingBucket, err := allPostingBucket.CreateBucketIfNotExists(encode64Bit(wordID))
			if err != nil {
				return err
			}

			docAlreadyExist := postingBucket.Get(encode64Bit(docID))
			if docAlreadyExist == nil {
				//If the document is not in the posting list before, increase DF by 1
				totalDFByte := postingBucket.Get(encode64Bit(0))
				var totalDF int32
				if totalDFByte != nil {
					//totalTerms stores the current total document frequency
					totalDF = decode32Bit(totalDFByte)
				} else {
					totalDF = 0
				}

				//Insert the total DF back
				err = postingBucket.Put(encode64Bit(0), encode32Bit(totalDF+1))
				if err != nil {
					return err
				}
			}

			err = postingBucket.Put(encode64Bit(docID), encode32Bit(termFreq))
			if err != nil {
				return err
			}

			termAlreadyExist := forwardBucket.Get(encode64Bit(wordID))
			if termAlreadyExist == nil {
				//Increase totalTerms by 1 if the term was not indexed for this document before
				totalTermsByte := forwardBucket.Get(encode64Bit(0))
				var totalTerms int32
				if totalTermsByte != nil {
					//totalTF stores the current total number of terms in the document
					totalTerms = decode32Bit(totalTermsByte)
				} else {
					totalTerms = 0
				}

				//Insert the total DF back
				err = forwardBucket.Put(encode64Bit(0), encode32Bit(totalTerms+1))

				if err != nil {
					return err
				}
			}

			if termFreq > maxTF {
				maxTF = termFreq
			}

			err = forwardBucket.Put(encode64Bit(wordID), encode32Bit(termFreq))
			if err != nil {
				return err
			}
		}

		//ID 0 stores the precomputed max TF
		err = forwardBucket.Put(encode64Bit(0), encode32Bit(maxTF))
		if err != nil {
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
			//The zeroth index has special meaning.
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

//GetDFOfTerm returns the document frequency of the term
func GetDFOfTerm(termID int64) int32 {
	var result int32
	postingDB.View(func(tx *bolt.Tx) error {
		allPostingBuc := tx.Bucket([]byte("posting"))
		specifiedPostingBuc := allPostingBuc.Bucket(encode64Bit(termID))
		if specifiedPostingBuc == nil {
			result = 0
		} else {
			returnByte := specifiedPostingBuc.Get(encode64Bit(0))
			result = decode32Bit(returnByte)
		}
		return nil
	})
	return result
}

//GetMaxTFOfDoc returns the maximum TF in the document
func GetMaxTFOfDoc(docID int64) int32 {
	var result int32
	postingDB.View(func(tx *bolt.Tx) error {
		allForwardBuc := tx.Bucket([]byte("forward"))
		specifiedForwardBuc := allForwardBuc.Bucket(encode64Bit(docID))
		if specifiedForwardBuc == nil {
			result = 0
		} else {
			returnByte := specifiedForwardBuc.Get(encode64Bit(0))
			result = decode32Bit(returnByte)
		}
		return nil
	})
	return result
}

//GetMaxTFOfDocStem returns the maximum TF in the document, used in the stemmed list
func GetMaxTFOfDocStem(docID int64) int32 {
	var result int32
	postingDB.View(func(tx *bolt.Tx) error {
		allForwardBuc := tx.Bucket([]byte("stem_forward"))
		specifiedForwardBuc := allForwardBuc.Bucket(encode64Bit(docID))
		if specifiedForwardBuc == nil {
			result = 0
		} else {
			returnByte := specifiedForwardBuc.Get(encode64Bit(0))
			result = decode32Bit(returnByte)
		}
		return nil
	})
	return result
}

//GetDocOfTerm returns a collection of documents containing the termID, together with their term frequency
func GetDocOfTerm(termID int64) (docIDCollection []int64, postingCollection []*Posting, total int32) {
	postingDB.View(func(tx *bolt.Tx) error {
		allPostingBuc := tx.Bucket([]byte("posting"))
		specifiedPostingBuc := allPostingBuc.Bucket(encode64Bit(termID))

		//Get the total document count in advance to save reallocation
		total = decode32Bit(specifiedPostingBuc.Get(encode64Bit(0)))
		docIDCollection = make([]int64, total)
		postingCollection = make([]*Posting, total)
		docCount := 0

		//Iterate through all the postings
		specifiedPostingBuc.ForEach(func(k, v []byte) error {
			id := decode64Bit(k)
			if id != 0 {
				docIDCollection[docCount] = id
				posting := decodePosting(v)
				postingCollection[docCount] = posting
				docCount++
			}
			return nil
		})
		return nil
	})
	return
}

//GetDocOfStemTerm returns a collection of documents containing the termID, together with their term frequency, used in stemmed list
func GetDocOfStemTerm(termID int64) (docIDCollection []int64, tfCollection []int32, total int32) {
	postingDB.View(func(tx *bolt.Tx) error {
		allPostingBuc := tx.Bucket([]byte("stem_posting"))
		specifiedPostingBuc := allPostingBuc.Bucket(encode64Bit(termID))

		//Get the total document count in advance to save reallocation
		total = decode32Bit(specifiedPostingBuc.Get(encode64Bit(0)))
		docIDCollection = make([]int64, total)
		tfCollection = make([]int32, total)
		docCount := 0

		//Iterate through all the postings
		specifiedPostingBuc.ForEach(func(k, v []byte) error {
			id := decode64Bit(k)
			if id != 0 {
				docIDCollection[docCount] = id
				tf := decode32Bit(v)
				tfCollection[docCount] = tf
				docCount++
			}
			return nil
		})
		return nil
	})
	return
}

//GetRootSquaredTermFreqOfDoc returns the length of document in the cosine similarity sense
func GetRootSquaredTermFreqOfDoc(docID int64) float64 {
	var sum int64
	sum = 0
	postingDB.View(func(tx *bolt.Tx) error {
		allForwardBuc := tx.Bucket([]byte("forward"))
		specifiedForwardBuc := allForwardBuc.Bucket(encode64Bit(docID))
		if specifiedForwardBuc == nil {
			sum = 0
		} else {
			specifiedForwardBuc.ForEach(func(k, v []byte) error {
				id := decode64Bit(k)
				if id != 0 {
					tf := decode32Bit(v)
					sum += int64(tf * tf)
				}
				return nil
			})
		}
		return nil
	})

	return math.Sqrt(float64(sum))
}

//TfIDPair is a pair of data
type TfIDPair struct {
	Tf int32
	Id int64
}

//GetRSStemTFOfDocAndTop5 returns the length of document in the cosine similarity sense AND the top 5 stemmed words, used in stemmed list
func GetRSStemTFOfDocAndTop5(docID int64) (rsSum float64, top5Pair []TfIDPair) {
	var sum int64
	sum = 0

	var pairSlice []TfIDPair
	postingDB.View(func(tx *bolt.Tx) error {
		allForwardBuc := tx.Bucket([]byte("stem_forward"))
		specifiedForwardBuc := allForwardBuc.Bucket(encode64Bit(docID))
		pairSlice = make([]TfIDPair, specifiedForwardBuc.Stats().KeyN)

		if specifiedForwardBuc == nil {
			sum = 0
		} else {
			index := 0
			specifiedForwardBuc.ForEach(func(k, v []byte) error {
				id := decode64Bit(k)
				if id != 0 {
					tf := decode32Bit(v)
					sum += int64(tf * tf)
					pairSlice[index].Tf = tf
					pairSlice[index].Id = id
				}
				index++
				return nil
			})
		}
		return nil
	})

	sort.Slice(pairSlice, func(i, j int) bool {
		//In descending order
		return pairSlice[i].Tf > pairSlice[j].Tf
	})

	if len(pairSlice) > 5 {
		top5Pair = pairSlice[:5]
	} else {
		top5Pair = pairSlice
	}

	rsSum = math.Sqrt(float64(sum))
	return
}
