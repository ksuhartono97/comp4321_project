package webserver

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"../../silver-rush/retrieval"
)

type UrlData struct {
	sourceUrl    string
	sourceID     int64
	foundUrl     []string
	pageTitle    string
	pageSize     int
	rawHTML      string
	lastModified int64
}

//var queryResult [1]UrlData = {UrlData{sourceUrl: "google.com", sourceID: "213", pageTitle:"Choco", pageSize:123, rawHtml:"lul", lastModified:"Yesterday"}}
var resultString = "Here is a string\n Thomas \n Dong \n Doo \n Dah\n"
var resultString2 = "Ding ding\n Dong\n \n Dudu"

type Page struct {
	Body      []byte
	StringArr []string
}

var expectedQueryResult []string

//Load up the result to the html page
func loadResult() *Page {
	//Construct an array for the result
	strRes := []string{}
	body := []byte(resultString)
	fmt.Println(body)

	//Decompose all the strings that are in the result
	for _, str := range expectedQueryResult {
		q := strings.Split(str, "\n")
		strRes = append(strRes, q...)
	}
	return &Page{Body: body, StringArr: strRes}
}

//Handler for the query page
func queryHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/query.html")
		t.Execute(w, nil)
	} else {
		//Instead of Println will export to something else here.
		r.ParseForm()
		// fmt.Println("Query:", r.Form["searchInput"])
		temp := strings.Join(r.Form["searchInput"], ",")

		//Submitting query to the search engine
		expectedQueryResult = retrieval.RetrieveRankedStringResult(temp)

		UpdateResultString(temp)
		http.Redirect(w, r, "/result", http.StatusSeeOther)
	}
}

//Handler for the result page
func resultHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("method:", r.Method) //get request method
	t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/results.html")
	p := loadResult()
	t.Execute(w, p)
}

//May be deprecated soon, will update later
func UpdateResultString(newString string) {
	resultString = newString
}

func StartWebServer() {
	expectedQueryResult = append(expectedQueryResult, resultString)
	expectedQueryResult = append(expectedQueryResult, resultString2)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/result", resultHandler)
	http.ListenAndServe(":8080", nil)
}
