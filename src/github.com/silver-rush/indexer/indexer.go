package indexer

import (
	"fmt"
	"sync"

	"strings"

	"../../../golang.org/x/net/html"
	"../../ksuhartono97/stopword_rmv"
	"../../reiver/go-porterstemmer"
	"../../silver-rush/database"
)

//Feed a page to the indexer
func Feed(docID int64, raw string, lastModify int64, size int32, parent int64, child []int64, title string) {
	//Map of words and term frequency.
	origMap := make(map[int64]database.Posting)
	stemMap := make(map[int64]int32)

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		fmt.Println("Error in html parsing", err)
		panic(err)
	}

	bodyNode := findBodyNode(doc)
	if bodyNode != nil {
		//Map is pass by reference, so we're cool.
		iterateNode(doc, origMap, stemMap, 0)
	} else {
		fmt.Println("Body not found.")
	}

	var wg sync.WaitGroup
	wg.Add(3)
	//Start goroutines to add words to the posting list
	go func(records map[int64]database.Posting) {
		defer wg.Done()
		database.BatchInsertIntoPostingList(docID, records)
	}(origMap)

	//Start goroutines to add words to the stemmed posting list
	go func(records map[int64]int32) {
		defer wg.Done()
		database.BatchInsertIntoStemmedList(docID, records)
	}(stemMap)

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
		d.ParentNum = 1
		d.Parent = make([]int64, 1)
		d.Parent[0] = parent
		d.Title = title
		database.InsertDocInfo(docID, &d)
	}()

	wg.Wait()
	fmt.Println("Indexed: ", docID, ". ")
}

//JustAddParentIDToURL does not re-index the page. It simply add one entry to the parentID of the current page.
func JustAddParentIDToURL(parentID, pageID int64) {
	d := database.GetDocInfo(pageID)
	if d != nil {
		for _, id := range d.Parent {
			if id == parentID {
				//If id already exist, go home.
				return
			}
		}
		d.ParentNum++
		d.Parent = append(d.Parent, parentID)
		database.InsertDocInfo(pageID, d)
	}
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

func tokenize(text string) (original, stemmed []string) {
	text = html.UnescapeString(text)
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
				lowercase := strings.ToLower(text[head:i])
				original = append(original, lowercase)
				if !stopword_rmv.CheckForStopword(lowercase) {
					stemmed = append(stemmed, porterstemmer.StemString(lowercase))
				}
				head = i + 1
			}
		}
	}

	//Deal with the last word
	if head != i {
		lowercase := strings.ToLower(text[head:i])
		original = append(original, lowercase)
		if !stopword_rmv.CheckForStopword(lowercase) {
			stemmed = append(stemmed, porterstemmer.StemString(lowercase))
		}
	}

	return original, stemmed
}

func iterateNode(node *html.Node, origMap map[int64]database.Posting, stemMap map[int64]int32, pos int32) {
	if node.Type == html.TextNode && node.Parent.Data != "script" && node.Parent.Data != "style" {
		original, stemmed := tokenize(html.UnescapeString(node.Data))

		//For the original queue
		if len(original) != 0 {
			//Collect word id using the word
			idList, _ := database.BatchGetIDWithWord(original)
			//fmt.Printf("%v\n", idList)
			for _, id := range idList {
				p := origMap[id]
				p.TermFreq++
				p.Positions = append(p.Positions, pos)
				origMap[id] = p

				pos++
			}
		}

		//For the stemmed queue
		if len(stemmed) != 0 {
			//Collect word id using the word
			idList, _ := database.BatchGetIDWithWord(stemmed)
			for _, id := range idList {
				stemMap[id]++
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		//Recursively iterate through all nodes
		iterateNode(child, origMap, stemMap, pos+10)
	}
}
