package retrieval

import (
	"regexp"
	"strings"
)

func analyzeQuery(query string) [][]string {
	doubleQuoteRegex := regexp.MustCompile("\"(.*?)\"")
	doubleQuoteText := doubleQuoteRegex.FindAllString(query, -1)
	var result [][]string

	//Phrase searches are marked by double quotes
	if doubleQuoteText != nil {
		for _, s := range doubleQuoteText {
			query = strings.Replace(query, s, "", -1)
			result = append(result, strings.Split(s[1:len(s)-1], " "))
		}
	}

	restOfTheQuery := strings.Split(query, " ")
	for _, s := range restOfTheQuery {
		//Each single word term will be a slice of 1 element
		v := make([]string, 1)
		v[0] = s
		result = append(result, v)
	}

	return result
}
