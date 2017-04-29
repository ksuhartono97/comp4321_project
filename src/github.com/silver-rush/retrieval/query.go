package retrieval

import (
	"strings"
)

func analyzeQuery(query string) []string {
	return strings.Split(query, " ")
}
