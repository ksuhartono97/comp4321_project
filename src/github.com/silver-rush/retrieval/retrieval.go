package retrieval

import (
	"math"
	"sort"

	"../../silver-rush/database"
)

//Retrieve will return the retrieved docID ranked with similarity
func Retrieve(query string) []int64 {
	queryTerms := analyzeQuery(query)

	totalDoc := database.GetTotalNumberOfDocument()

	type tfIdfStruct struct {
		docID int64
		tfIdf float64
	}

	tfIdfChannel := make(chan tfIdfStruct)
	doneChannel := make(chan bool)
	for _, group := range queryTerms {
		//Group is a slice. Contains more than one value only in phrase searches
		if len(group) == 1 {
			//Single word
			go func(s string) {
				//Compute idf first
				termID, exist := database.GetIDWithWordDoNotCreate(s)
				if !exist {
					//Zero term weight
				} else {
					docIDCollection, termFreqCollection, documentFreq := database.GetDocOfTerm(termID)
					inverseDocFreq := math.Log2(float64(totalDoc) / float64(documentFreq))

					for i := 0; i < int(documentFreq); i++ {
						//This result is not divided by max tf yet. Will do afterwards.
						tfIdfChannel <- tfIdfStruct{docIDCollection[i], float64(termFreqCollection[i]) * inverseDocFreq}
					}
				}
				doneChannel <- true
			}(group[0])
		} else {
			//Phrase search.
		}
	}

	totalTfIdfMap := make(map[int64]float64)
	for c := 0; c < len(queryTerms); {
		//Collect result here
		select {
		case r := <-tfIdfChannel:
			totalTfIdfMap[r.docID] += r.tfIdf
		case <-doneChannel:
			c++

		}
	}

	type similarityStruct struct {
		docID      int64
		similarity float64
	}
	similaritySlice := make([]similarityStruct, len(totalTfIdfMap))
	i := 0
	queryLength := len(queryTerms)
	for k, v := range totalTfIdfMap {
		docLength := database.GetRootSquaredTermFreqOfDoc(k)
		similaritySlice[i].docID = k
		similaritySlice[i].similarity = v / docLength / float64(queryLength)
		i++
	}

	//Sort it so that the array is in the order of similarity
	sort.Slice(similaritySlice, func(i, j int) bool {
		return similaritySlice[i].similarity < similaritySlice[j].similarity
	})

	rankedResult := make([]int64, len(similaritySlice))
	for i, s := range similaritySlice {
		rankedResult[i] = s.docID
	}

	return rankedResult
}
