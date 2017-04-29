package webserver

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
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
var resultString = "Here is a string\n Thomas \n Dong \n Doo \n Dah"

type Page struct {
	Body  []byte
	StringArr []string
}


func loadResult() (*Page) {
	// filename := "resources/" + title + ".txt"
	body := []byte(resultString)
	fmt.Println(body);
	q := strings.Split(resultString, "\n")
	// p.StringArr = q
	for _, content := range q {
		fmt.Println(content)
	}
	return &Page{Body: body, StringArr:q}
}

func queryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/query.html")
		t.Execute(w, nil)
	} else {
    //Instead of Println will export to something else here.
		r.ParseForm()
		fmt.Println("Query:", r.Form["searchInput"])
		temp := strings.Join(r.Form["searchInput"], ",")
		UpdateResultString(temp)
    http.Redirect(w, r, "/result", http.StatusSeeOther)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/results.html")
	p := loadResult()
	t.Execute(w, p)
}

func UpdateResultString (newString string) {
	resultString = newString
}

func StartWebServer() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/result", resultHandler)
	http.ListenAndServe(":8080", nil)
}
