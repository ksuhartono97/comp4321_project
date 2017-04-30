This program require Go to build. Tested on lab2 machine. The $GOPATH has to be set to the directory of this project, which can be done is bash using "export GOPATH=..."

1. Running the crawler and indexer. Go into /src, then type "go run main.go"

2. Reading and printing the database. Go into /src, then type "go test"

Sorry for any extra troubles caused by this procedure.

The database file are located in /src/db/
The spider_result.txt is in /src/

To open the website, after doing "go run main.go" in /src, open a
web browser and go to http://localhost:8080/query

SPECIAL NOTE: if during runtime you encounter errors saying socket: too many open files
that causes the crawler to crash, PLEASE TYPE IN THIS COMMAND:

ulimit -n 2000
