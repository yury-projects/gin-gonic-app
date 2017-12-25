package rss

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"net/http"
)

const rss_feed = "https://rss.nytimes.com/services/xml/rss/nyt/Americas.xml"

type GUID struct {
	gorm.Model
	LastID string
}

func GetListOfNewGUIDs() []string {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(rss_feed)

	var db_guid GUID

	db := Database()

	db.First(&db_guid)

	counter := 0

	var guids []string = make([]string, len(feed.Items))

	for ind, item := range feed.Items {
		fmt.Println(item.GUID)
		if item.GUID != db_guid.LastID {
			guids[ind] = item.GUID
			counter++
		} else {
			break
		}
	}

	if 0 == counter {
		guids = []string{}
	} else {
		guids = guids[0:counter]

		db_guid.LastID = guids[0]
		db.Save(&db_guid)
	}

	return guids
}

func CheckFeedAndNotify() {
	guids := GetListOfNewGUIDs()

	fmt.Println(guids)

	if len(guids) > 0 {
		NotifyNewContent(guids)
	}
}

func GetLatestRssFeed(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"objects": GetListOfNewGUIDs()})
}
