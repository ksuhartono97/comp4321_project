package indexer

import (
	"io"
	"time"
)

//Feed a page to the indexer, the indexer will reroute the tasks to different goroutines
func Feed(docID int, reader io.ReadCloser, lastModify time.Time, size int, parent int, child []int) {

}
