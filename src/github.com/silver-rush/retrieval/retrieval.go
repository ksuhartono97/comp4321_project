package retrieval

import (
	"math"
	"sort"

	"fmt"

	"../../silver-rush/database"
)

//RetrieveRankedStringResult returns readable string result
func RetrieveRankedStringResult(query string) []string {
	docIDSlice := RetrieveRankedDocID(query)
	fmt.Printf("Retrieval size: %d\n", len(docIDSlice))
	allResult := make([]string, len(docIDSlice))
	for i, id := range docIDSlice {
		docInfo := database.GetDocInfo(id)
		url := database.GetURLWithID(id)
		pageResult := fmt.Sprintf("%s\n <b>URL</b>: %s \nSize: %d \nTime: %d\n\n",
			docInfo.Title,
			url,
			docInfo.Size,
			docInfo.Time)
		allResult[i] = pageResult
	}
	return allResult
}

//RetrieveRankedDocID will return the retrieved docID ranked with similarity
func RetrieveRankedDocID(query string) []int64 {
	queryTerms := analyzeQuery(query)
	fmt.Printf("Query size: %d\n", len(queryTerms))
	fmt.Printf("Query terms: %v\n", queryTerms)

	totalDoc := database.GetTotalNumberOfDocument()

	type tfIdfStruct struct {
		docID int64
		tfIdf float64
	}

	tfIdfChannel := make(chan tfIdfStruct)
	doneChannel := make(chan bool)
	for _, group := range queryTerms {
		//Group is a slice. Contains more than one value only in phrase searches
		fmt.Printf("Group: %v Len: %d\n", group, len(group))
		if len(group) == 1 {
			//Single word
			go func(s string) {
				//Compute idf first
				fmt.Printf("Term: %s\n", s)
				termID, exist := database.GetIDWithWordDoNotCreate(s)
				if !exist {
					fmt.Printf("Do not exist.\n")
					//Zero term weight
				} else {
					docIDCollection, postingCollection, documentFreq := database.GetDocOfTerm(termID)
					inverseDocFreq := math.Log2(float64(totalDoc) / float64(documentFreq))

					for i := 0; i < int(documentFreq); i++ {
						//This result is not divided by max tf yet. Will do afterwards.
						tfIdfChannel <- tfIdfStruct{docIDCollection[i], float64(postingCollection[i].TermFreq) * inverseDocFreq}
					}
				}
				doneChannel <- true
			}(group[0])
		} else {
			//Phrase search.
			go func(group []string) {
				termIDSlice := make([]int64, len(group))
				allExist := true
				var minDF int32
				var indexWithMinDF int
				allDocIDCollection := make([][]int64, len(group))
				allPostingCollection := make([][]*database.Posting, len(group))
				for i, term := range group {
					id, exist := database.GetIDWithWordDoNotCreate(term)
					if !exist {
						//If one of the term in the query does not exist, give up immediately
						allExist = false
						break
					}
					termIDSlice[i] = id
					docIDCollection, postingCollection, documentFreq := database.GetDocOfTerm(id)
					allDocIDCollection[i] = docIDCollection
					allPostingCollection[i] = postingCollection
					//Assign the first term to be the min. DF first no matter what
					if i == 0 || documentFreq < minDF {
						minDF = documentFreq
						indexWithMinDF = i
					}
				}

				fmt.Printf("IDs: %v Len: %d\n", termIDSlice, len(group))
				//This specifies the amount of offset the term should have in the phrase
				positionOffset := make([]int, len(group))
				for i := range group {
					positionOffset[i] = i - indexWithMinDF
				}

				progressPointers := make([]int, len(group))
				if allExist {
					dfOfPhrase := 0
					tfMap := make(map[int64]int)
					for i, docID := range allDocIDCollection[indexWithMinDF] {
						tfInDoc := 0
						for _, startPos := range allPostingCollection[indexWithMinDF][i].Positions {
							documentPossible := true
							for j := range group {
								positionPossible := false
								if j != indexWithMinDF {
									for (len(allDocIDCollection[j]) < progressPointers[j]) && (allDocIDCollection[j][progressPointers[j]] < docID) {
										//Skip until the end or overshoot
										progressPointers[j]++
									}

									if len(allDocIDCollection[j]) < progressPointers[j] && (allDocIDCollection[j][progressPointers[j]] == docID) {
										//Look into the position if docID matches
										for _, tarPos := range allPostingCollection[progressPointers[j]][i].Positions {
											if int(tarPos) == int(startPos)+positionOffset[j] {
												positionPossible = true
												break
											}
										}
										//If one of the term does not satisfy, go search for next position
										if !positionPossible {
											break
										}
									} else {
										//This documnet is not possible to have a match. Skip the rest of the positions.
										documentPossible = false
									}
								}
								if !documentPossible {
									//Go to next docID
									break
								}

								if positionPossible {
									//If position is still possible after all searches, it exists for real
									tfInDoc++
								}
							}
						}
						if tfInDoc != 0 {
							tfMap[docID] = tfInDoc
							dfOfPhrase++
						}
					}

					//After everything is computed, get tfidf (as we can only get DF after all searches)
					inverseDocFreq := math.Log2(float64(totalDoc) / float64(dfOfPhrase))
					for k, v := range tfMap {
						tfIdfChannel <- tfIdfStruct{k, float64(v) * inverseDocFreq}
					}
				}
				doneChannel <- true
			}(group)
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
		//Obtain cosine similarity
		docLength := database.GetRootSquaredTermFreqOfDoc(k)
		maxTF := database.GetMaxTFOfDoc(k)
		similaritySlice[i].docID = k
		similaritySlice[i].similarity = v / docLength / float64(queryLength) / float64(maxTF)
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
