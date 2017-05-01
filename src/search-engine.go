package main

import (
	"./github.com/ksuhartono97/stopword_rmv"
	"./github.com/ksuhartono97/webserver"
	"./github.com/silver-rush/database"
)

func main() {
	stopword_rmv.ConstructRegex()
	database.OpenAllDatabaseReadOnly()
	defer database.CloseAllDatabase()
	webserver.StartWebServer()
}
