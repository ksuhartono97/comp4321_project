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
	origMap := make(map[int64]*database.Posting)
	stemMap := make(map[int64]int32)

	doc, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		fmt.Println("Error in html parsing", err)
		panic(err)
	}

	//Put the title inside an inverted file
	database.InsertTitleForDoc(docID, title)

	bodyNode := findBodyNode(doc)
	if bodyNode != nil {
		//Map is pass by reference, so we're cool.
		iterateNode(doc, origMap, stemMap, 50)
	} else {
		fmt.Println("Body not found.")
	}

	var wg sync.WaitGroup
	wg.Add(4)
	//Start goroutines to add words to the posting list
	go func(records map[int64]*database.Posting) {
		defer wg.Done()
		database.BatchInsertIntoPostingList(docID, records)
	}(origMap)

	//Start goroutines to add words to the stemmed posting list
	go func(records map[int64]int32) {
		defer wg.Done()
		database.BatchInsertIntoStemmedList(docID, records)
	}(stemMap)

	//Also put the title into normal posting list
	go func() {
		defer wg.Done()
		title = strings.ToLower(title)
		titleTerms := strings.Split(title, " ")
		var stemmed []string
		for _, term := range titleTerms {
			if !stopword_rmv.CheckForStopword(term) {
				stemmed = append(stemmed, porterstemmer.StemString(term))
			}
		}

		normalIDs, _ := database.BatchGetIDWithWord(titleTerms)
		stemmedIDs, _ := database.BatchGetIDWithWord(stemmed)

		postings := make(map[int64]*database.Posting)
		for i, termID := range normalIDs {
			if postings[termID] != nil {
				postings[termID].TermFreq += 3
				postings[termID].Positions = append(postings[termID].Positions, int32(i))
			} else {
				var p database.Posting
				p.TermFreq = 5
				pos := make([]int32, 1)
				pos[0] = int32(i)
				p.Positions = pos
				postings[termID] = &p
			}
		}

		stemedRecord := make(map[int64]int32)
		for _, termID := range stemmedIDs {
			stemedRecord[termID] += 4
		}

		database.BatchInsertIntoPostingList(docID, postings)
		database.BatchInsertIntoStemmedList(docID, stemedRecord)
	}()

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
		if parent > 0 {
			d.ParentNum = 1
			d.Parent = make([]int64, 1)
			d.Parent[0] = parent
		} else {
			d.ParentNum = 0
			d.Parent = make([]int64, 0)
		}

		d.Title = title
		database.InsertDocInfo(docID, &d)
	}()

	wg.Wait()
	fmt.Printf("Indexed: %d with parent: %d", docID, parent)
}

//JustAddParentIDToURL does not re-index the page. It simply add one entry to the parentID of the current page.
func JustAddParentIDToURL(parentID, pageID int64) {
	d := database.GetDocInfo(pageID)
	if d != nil && parentID != pageID {
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
	text = strings.ToLower(text)
	head := 0

	i := 0
	for ; i < len(text); i++ {
		//Index english alphahets only
		if (text[i] < 'a' || text[i] > 'z') && (text[i] < 'A' || text[i] > 'Z') {
			if i == head {
				head++
			} else {
				//Append a slice
				original = append(original, text[head:i])
				if !stopword_rmv.CheckForStopword(text[head:i]) {
					stemmed = append(stemmed, porterstemmer.StemString(text[head:i]))
				}
				head = i + 1
			}
		}
	}

	//Deal with the last word
	if head != i {
		original = append(original, text[head:i])
		if !stopword_rmv.CheckForStopword(text[head:i]) {
			stemmed = append(stemmed, porterstemmer.StemString(text[head:i]))
		}
	}

	return original, stemmed
}

func iterateNode(node *html.Node, origMap map[int64]*database.Posting, stemMap map[int64]int32, pos int32) {
	if node.Type == html.TextNode && node.Parent.Data != "script" && node.Parent.Data != "style" {

		//Giving extra attention to tagged terms
		emphasisPower := 1
		if node.Parent.Data == "h1" || node.Parent.Data == "h2" || node.Parent.Data == "h3" {
			emphasisPower = 3
		}

		if node.Parent.Data == "h4" || node.Parent.Data == "h5" || node.Parent.Data == "h6" {
			emphasisPower = 2
		}

		if node.Parent.Data == "b" || node.Parent.Data == "i" || node.Parent.Data == "u" {
			emphasisPower = 2
		}

		original, stemmed := tokenize(html.UnescapeString(node.Data))

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			//For the original queue
			if len(original) != 0 {
				//Collect word id using the word
				idList, _ := database.BatchGetIDWithWord(original)
				//fmt.Printf("%v\n", idList)
				for _, id := range idList {
					p := origMap[id]
					if p == nil {
						var posting database.Posting
						p = &posting
					}
					p.TermFreq += int32(emphasisPower)
					p.Positions = append(p.Positions, pos)
					origMap[id] = p

					pos++
				}
			}
		}()

		go func() {
			defer wg.Done()
			//For the stemmed queue
			if len(stemmed) != 0 {
				//Collect word id using the word
				idList, _ := database.BatchGetIDWithWord(stemmed)
				for _, id := range idList {
					stemMap[id] += int32(emphasisPower)
				}
			}
		}()

		wg.Wait()

	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		//Recursively iterate through all nodes
		iterateNode(child, origMap, stemMap, pos+10)
	}
}
