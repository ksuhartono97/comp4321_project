package webserver

import (
	"html/template"
	"net/http"
	"strings"
	"fmt"
	"time"

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
var resultString = "Here is a string\n Dong \n Doo \n Dah\n"
var resultString2 = "Ding ding\n Dong\n \n Dudu"

type Page struct {
	Body      []byte
	Data			string
}

var expectedQueryResult []string

//Load up the result to the html page
func loadResult() *Page {
	//Construct an array for the result
	body := []byte(resultString)

	var tempString string = ""
	//Decompose all the strings that are in the result
	for _, str := range expectedQueryResult {
		replacedBR := strings.Replace(str,"\n","<br>",-1)
		tempString += replacedBR
	}
	return &Page{Body: body, Data: tempString}
}

//Handler for the query page
func queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/query.html")
		t.Execute(w, nil)
	} else {

		//Parsing contents of the form
		r.ParseForm()
		temp := strings.Join(r.Form["searchInput"], ",")

		start := time.Now()
		//Submitting query to the search engine
		expectedQueryResult = retrieval.RetrieveRankedStringResult(temp)
		elapsed := time.Since(start)
		fmt.Println("Query took ", elapsed)

		http.Redirect(w, r, "/result", http.StatusSeeOther)
	}
}

//Handler for the result page
func resultHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/results.html")
	html := template.HTML(loadResult().Data)
	t.Execute(w, map[string]interface{}{
   "Body": html,
	})
}

//Handler for /
func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/query", http.StatusSeeOther)
}

//Run and host the webserver so that it will be accessible through "localhost:8080"
func StartWebServer() {
	expectedQueryResult = append(expectedQueryResult, resultString)
	expectedQueryResult = append(expectedQueryResult, resultString2)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/result", resultHandler)
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
