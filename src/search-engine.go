package main

import (
	"./github.com/ksuhartono97/webserver"
	"./github.com/silver-rush/database"
)

func main() {
	database.OpenAllDatabaseReadOnly()
	defer database.CloseAllDatabase()
	webserver.StartWebServer()
}
