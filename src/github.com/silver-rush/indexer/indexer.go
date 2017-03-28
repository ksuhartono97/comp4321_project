package indexer

import (
	"fmt"
	"io"
	"time"

	"strings"

	"unicode"

	"golang.org/x/net/html"
)

//Feed a page to the indexer
func Feed(docID int, reader io.ReadCloser, lastModify time.Time, size int, parent int, child []int) {
	//Map of words and term frequency.
	//TODO: Change string to word ID
	wordMap := make(map[string]int)

	doc, err := html.Parse(reader)
	if err != nil {
		fmt.Println("Error in parsing")
	}

	//Map is pass by reference, so we're cool
	iterateNode(doc, wordMap)
}

var replacer = strings.NewReplacer("\n", " ", "\t", " ")

func tokenize(text *string) []string {
	*text = html.UnescapeString(*text)
	*text = replacer.Replace(*text)
	var tokens []string
	head := 0
	for i, c := range *text {
		if !unicode.IsLetter(c) {
			if i == head {
				//fmt.Printf("B1: %d %d %d\n", head, i, c)
				head++
			} else {
				tokens = append(tokens, (*text)[head:i])
				//fmt.Printf("B2: %d %d %d\n", head, i, c)
				head = i + 1
			}
		}
	}
	return tokens
}

func iterateNode(node *html.Node, wordMap map[string]int) {
	if node.Type == html.TextNode && node.Parent.Data != "script" {
		list := tokenize(&(node.Data))
		for _, s := range list {
			fmt.Printf("%s\n", strings.Replace(s, "\xfffd", "", -1))
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		iterateNode(child, wordMap)
	}
}
