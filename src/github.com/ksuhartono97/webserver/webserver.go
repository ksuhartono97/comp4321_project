package webserver

import (
	"fmt"
	"html/template"
	"net/http"
)

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
	t.Execute(w, nil)
}

func StartWebServer() {
	// http.Handle("/", http.FileServer(http.Dir("./html")))
	http.HandleFunc("/query", queryHandler)
	http.HandleFunc("/result", resultHandler)
	http.ListenAndServe(":8080", nil)
}
