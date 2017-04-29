package retrieval

//Retrieve will returns the retrieved, ranked result from the database
func Retrieve(query string) []string {
	queryTerms := analyzeQuery(query)
	return queryTerms
}
