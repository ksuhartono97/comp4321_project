package main

import (
	"./github.com/ksuhartono97/stopword_rmv"
)

//Only for sample delete later!! Note, the directory of the stopword_list might have to be changed
func main() {
	stopword_rmv.ConstructRegex()
  fmt.Println(stopword_rmv.CheckForStopword("about"))
  fmt.Println(stopword_rmv.CheckForStopword("time"))
}
