package indexer

import "../../silver-rush/database"

//CheckURL will check whether you should crawl this page or not. Returns a valid document ID only if shouldCrawl is true.
func CheckURL(url string, lastModify int64) (id int64, shouldCrawl bool) {
	id = -1
	shouldCrawl = false
	id, created := database.GetURLID(url)
	if created {
		//Must crawl if url does not exist in database
		return id, true
	}

	//Otherwise, check the last modified date
	docInfo := database.GetDocInfo(id)
	if docInfo.Time < lastModify {
		shouldCrawl = true
	}

	return
}
