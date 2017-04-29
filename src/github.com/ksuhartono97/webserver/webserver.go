package webserver

import (
	"fmt"
	"html/template"
	"net/http"
	"io/ioutil"
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

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := "resources/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	fmt.Println(body);
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
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
    http.Redirect(w, r, "/result", http.StatusSeeOther)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	t, _ := template.ParseFiles("./github.com/ksuhartono97/webserver/html/results.html")
	p, _ := loadPage("queryResult")
	t.Execute(w, p)
}

func StartWebServer() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/result", resultHandler)
	http.ListenAndServe(":8080", nil)
}
