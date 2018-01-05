package slack

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mmcdole/gofeed"
	"github.com/yury-projects/gin-gonic-app/database"
)

const rssFeed = "https://rss.nytimes.com/services/xml/rss/nyt/Americas.xml"

type RSSCommand struct {
	CommandInterface
}

type GUID struct {
	gorm.Model
	LastID string
}

func (rssCommand *RSSCommand) IsValidText(text string) bool {

	return true
}

func (rssCommand *RSSCommand) GetCommandResponse() interface{} {

	rssItems := GetListOfNewRSSItems(9)

	if len(rssItems) == 0 || rssItems[0] == nil {
		return map[string]string{"text": "Nothing new to show here"}
	}

	var attachments []map[string]interface{}

	for _, rssItem := range rssItems {
		attachment := map[string]interface{}{
			"text":        rssItem.Description,
			"title":       rssItem.Title,
			"title_link":  rssItem.Link,
			"author_icon": "https://static01.nyt.com/images/icons/t_logo_291_black.png",
		}

		if rssItem.Image != nil {
			attachment["image_url"] = rssItem.Image.URL
		} else {
			// Trying extensions media -> content[0] -> attrs -> url
			media, exists := rssItem.Extensions["media"]

			if exists {
				content, exists := media["content"]

				if exists {
					attachment["thumb_url"] = content[0].Attrs["url"]
				}
			}
		}

		if rssItem.Author != nil {
			attachment["author_name"] = "By " + strings.Title(strings.ToLower(rssItem.Author.Name))
		}

		if rssItem.PublishedParsed != nil {
			attachment["ts"] = rssItem.PublishedParsed.Unix()
		}

		attachments = append(attachments, attachment)
	}

	return map[string]interface{}{"attachments": attachments}
}

// GetListOfNewRSSItems -
func GetListOfNewRSSItems(count int) []*gofeed.Item {
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL(rssFeed)

	var dbGUID GUID

	db := database.Database()

	db.First(&dbGUID)

	counter := 0

	if count > 0 {

		return feed.Items[:count]
	}

	var listOfItems = make([]*gofeed.Item, len(feed.Items))

	for ind, item := range feed.Items {
		fmt.Println(item.GUID)
		if item.GUID != dbGUID.LastID {
			listOfItems[ind] = item
			counter++
		} else {
			break
		}
	}

	if 0 == counter {
		listOfItems = make([]*gofeed.Item, 1)
	} else {
		listOfItems = listOfItems[0:counter]

		dbGUID.LastID = listOfItems[0].GUID
		db.Save(&dbGUID)
	}

	return listOfItems
}

func CheckFeedAndNotify() {
	guids := GetListOfNewRSSItems(-1)

	fmt.Println(guids)

	if len(guids) > 0 {
		// NotifyNewContent(guids)
		fmt.Println(guids)
	}
}

func GetLatestRssFeed(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{"objects": GetListOfNewRSSItems(9)})
}
