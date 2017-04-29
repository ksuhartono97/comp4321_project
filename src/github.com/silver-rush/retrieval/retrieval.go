package retrieval

//Retrieve will returns the retrieved, ranked result from the database
func Retrieve(query string) [][]string {
	queryTerms := analyzeQuery(query)

	// for _, group := range queryTerms {
	// 	//Group is a slice. Contains more than one value only in phrase searches
	// }

	return queryTerms
}
