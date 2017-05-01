package stopword_rmv

import (
  "io/ioutil"
  "fmt"
  "strings"
  "regexp"
)

var reStr=""

//Match regex with word to check for stopwords
func CheckForStopword(word string) (shouldRemove bool) {
  regex := regexp.MustCompile(reStr)
  shouldRemove = regex.MatchString(word)
  return
}

//Call this function once, to construct the regex
func ConstructRegex() {
  content, err := ioutil.ReadFile("resources/stopword_list.txt")
  if err != nil {
      fmt.Println(err)
  }
  reStr ="^("

  words := strings.Split(string(content), "\n")
  for i, word := range words {
      if i != 0 {
          reStr += `|`
      }
      reStr += word
  }
  reStr += ")$"
}
