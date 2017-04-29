package indexer

import (
	"fmt"
	"sync"

	"strings"

	"../../../golang.org/x/net/html"
	"../../silver-rush/database"
)

//Feed a page to the indexer
func Feed(docID int64, raw string, lastModify int64, size int32, parent int64, child []int64, title string) {
	//Map of words and term frequency.
	wordMap := make(map[int64]database.Posting)

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		fmt.Println("Error in html parsing", err)
		panic(err)
	}

	bodyNode := findBodyNode(doc)
	if bodyNode != nil {
		//Map is pass by reference, so we're cool.
		iterateNode(doc, wordMap, 0)
	} else {
		fmt.Println("Body not found.")
	}

	var wg sync.WaitGroup
	wg.Add(2)
	//Start goroutines to add words to the posting list
	go func(records map[int64]database.Posting) {
		defer wg.Done()
		database.BatchInsertIntoPostingList(docID, records)
	}(wordMap)

	go func() {
		defer wg.Done()
		var d database.DocInfo

		//Make sure the children are unique
		childMap := make(map[int64]bool)
		var uniqueChild []int64
		for _, id := range child {
			if childMap[id] == false {
				childMap[id] = true
				uniqueChild = append(uniqueChild, id)
			}
		}

		d.ChildNum = int32(len(uniqueChild))
		d.Child = uniqueChild

		d.Size = size
		d.Time = lastModify
		d.ParentID = parent
		d.Title = title
		database.InsertDocInfo(docID, &d)
	}()

	wg.Wait()
	fmt.Println("Indexed: ", docID, ". ")
}

func findBodyNode(node *html.Node) *html.Node {
	if node.Type == html.ElementNode && node.Data == "body" {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		//Recursively iterate through all nodes
		returnedNode := findBodyNode(child)
		if returnedNode != nil {
			return returnedNode
		}
	}
	return nil
}

func tokenize(text string) []string {
	text = html.UnescapeString(text)
	var tokens []string
	head := 0

	//If I obtain the i for range, some indexes are skipped. No idea why.
	i := 0
	for ; i < len(text); i++ {
		//Index english alphahets only
		if (text[i] < 'a' || text[i] > 'z') && (text[i] < 'A' || text[i] > 'Z') {
			if i == head {
				head++
			} else {
				//Append a slice
				tokens = append(tokens, text[head:i])
				head = i + 1
			}
		}
	}

	//Deal with the last word
	if head != i {
		tokens = append(tokens, text[head:i])
	}

	return tokens
}

func iterateNode(node *html.Node, wordMap map[int64]database.Posting, pos int32) {
	if node.Type == html.TextNode && node.Parent.Data != "script" && node.Parent.Data != "style" {
		wordList := tokenize(html.UnescapeString(node.Data))
		if len(wordList) != 0 {
			//Collect word id using the word
			idList, _ := database.BatchGetIDWithWord(wordList)
			//fmt.Printf("%v\n", idList)
			for _, id := range idList {
				p := wordMap[id]
				p.TermFreq++
				p.Positions = append(p.Positions, pos)
				wordMap[id] = p

				pos++
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		//Recursively iterate through all nodes
		iterateNode(child, wordMap, pos+10)
	}
}
