package indexer

import "../../silver-rush/database"

//CheckURL will check whether you should crawl this page or not.
func CheckURL(url string, lastModify int64) (id int64, shouldCrawl bool) {
	shouldCrawl = true
	id, created := database.GetURLID(url)
	if created {
		//Must crawl if url does not exist in database
		return id, true
	}

	//Otherwise, check the last modified date
	docInfo := database.GetDocInfo(id)
	if docInfo != nil {
		if docInfo.Time >= lastModify {
			shouldCrawl = false
		}
	}

	return
}
