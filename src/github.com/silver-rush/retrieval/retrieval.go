package retrieval

import (
	"math"
	"sort"

	"fmt"

	"../../reiver/go-porterstemmer"
	"../../silver-rush/database"
)

//Result is a struct containing the result of retrieval
type Result struct {
	docID int64
	score float64
	top5  []database.TfIDPair
}

//RetrieveRankedStringResult returns readable string result
func RetrieveRankedStringResult(query string) []string {
	docIDSlice := RetrieveRankedDocID(query)
	fmt.Printf("Retrieval size: %d\n", len(docIDSlice))
	allResult := make([]string, len(docIDSlice))
	for i, result := range docIDSlice {
		docInfo := database.GetDocInfo(result.docID)
		url := database.GetURLWithID(result.docID)
		pageResult := fmt.Sprintf("%s\n <b>URL</b>: <a href=\"%s\">%s</a> \nSize: %d \nTime: %d\n\n",
			docInfo.Title,
			url,
			url,
			docInfo.Size,
			docInfo.Time)
		allResult[i] = pageResult
	}
	return allResult
}

//RetrieveRankedDocID will return the retrieved docID ranked with similarity
func RetrieveRankedDocID(query string) []Result {
	queryTerms := analyzeQuery(query)
	fmt.Printf("Query size: %d\n", len(queryTerms))
	fmt.Printf("Query terms: %v\n", queryTerms)

	totalDoc := database.GetTotalNumberOfDocument()

	type tfIdfStruct struct {
		docID int64
		tfIdf float64
	}

	origTfIdfChannel := make(chan tfIdfStruct)
	stemTfIdfChannel := make(chan tfIdfStruct)
	doneChannel := make(chan bool)
	totalSearches := 0
	for _, group := range queryTerms {
		//Group is a slice. Contains more than one value only in phrase searches
		fmt.Printf("Group: %v Len: %d\n", group, len(group))
		if len(group) == 1 {
			//Single word
			totalSearches += 2 //Stemmed and unstemmed

			go func(s string) { //This one search in the unstemmed list
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
						origTfIdfChannel <- tfIdfStruct{docIDCollection[i], float64(postingCollection[i].TermFreq) * inverseDocFreq}
					}
				}
				doneChannel <- true
			}(group[0])

			go func(s string) { //This one search in the stemmed list
				//Compute idf first
				s = porterstemmer.StemString(s)
				fmt.Printf("Term: %s\n", s)
				termID, exist := database.GetIDWithWordDoNotCreate(s)
				if !exist {
					fmt.Printf("Do not exist.\n")
					//Zero term weight
				} else {
					docIDCollection, tfCollection, documentFreq := database.GetDocOfStemTerm(termID)
					inverseDocFreq := math.Log2(float64(totalDoc) / float64(documentFreq))

					for i := 0; i < int(documentFreq); i++ {
						//This result is not divided by max tf yet. Will do afterwards.
						stemTfIdfChannel <- tfIdfStruct{docIDCollection[i], float64(tfCollection[i]) * inverseDocFreq}
					}
				}
				doneChannel <- true
			}(group[0])

		} else {
			//Phrase search.
			totalSearches++ //Unstemmed only
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
				mapOfIDAndPosition := make([]map[int64]*database.Posting, len(group))
				for i := range group {
					positionOffset[i] = i - indexWithMinDF
					if i != indexWithMinDF {
						mapOfIDAndPosition[i] = make(map[int64]*database.Posting)
						for j, docID := range allDocIDCollection[i] {
							mapOfIDAndPosition[i][docID] = allPostingCollection[i][j]
						}
					} else {
						mapOfIDAndPosition[i] = nil
					}
				}

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

									if mapOfIDAndPosition[j][docID] != nil {
										//Look into the position if docID matches
										for _, tarPos := range mapOfIDAndPosition[j][docID].Positions {
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
						origTfIdfChannel <- tfIdfStruct{k, float64(v) * inverseDocFreq}
					}
				}
				doneChannel <- true
			}(group)
		}
	}

	origTotalTfIdfMap := make(map[int64]float64)
	stemTotalTfIdfMap := make(map[int64]float64)
	for c := 0; c < totalSearches; {
		//Collect result here
		select {
		case r := <-origTfIdfChannel:
			origTotalTfIdfMap[r.docID] += r.tfIdf
		case r := <-stemTfIdfChannel:
			stemTotalTfIdfMap[r.docID] += r.tfIdf
		case <-doneChannel:
			c++

		}
	}

	totalSimilarityMap := make(map[int64]float64)
	top5StemmedMap := make(map[int64][]database.TfIDPair)
	queryLength := len(queryTerms)

	for docID, tfIdf := range stemTotalTfIdfMap {
		var docLength float64
		docLength, top5StemmedMap[docID] = database.GetRSStemTFOfDocAndTop5(docID)
		maxTF := database.GetMaxTFOfDocStem(docID)
		totalSimilarityMap[docID] = tfIdf / docLength / float64(queryLength) / float64(maxTF)
	}

	for docID, tfIdf := range origTotalTfIdfMap {
		docLength := database.GetRootSquaredTermFreqOfDoc(docID)
		maxTF := database.GetMaxTFOfDoc(docID)
		totalSimilarityMap[docID] = tfIdf / docLength / float64(queryLength) / float64(maxTF)

		if top5StemmedMap[docID] == nil {
			//If top 5 have not been obtained, obtain them now!
			_, top5StemmedMap[docID] = database.GetRSStemTFOfDocAndTop5(docID)
		}
	}

	resultSlice := make([]Result, len(totalSimilarityMap))
	i := 0
	for k, v := range totalSimilarityMap {
		resultSlice[i].docID = k
		resultSlice[i].score = v
		resultSlice[i].top5 = top5StemmedMap[k]
		i++
	}

	//Sort it so that the array is in the order of similarity
	sort.Slice(resultSlice, func(i, j int) bool {
		//In descending order
		return resultSlice[i].score > resultSlice[j].score
	})

	return resultSlice
}
