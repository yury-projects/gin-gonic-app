package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/gin-gonic/gin/json"
	"bytes"
	"strings"
	"os"
)

const header_content_type = "application/json"

var (
	incoming_webhook_url = os.Getenv("GIN_GONIC_WEBHOOK_URL")
)

func NotifyNewContent(list_of_urls []string) {

	for i := len(list_of_urls) - 1; i >= 0; i-- {

		json_value, _ := json.Marshal(map[string]string{"text": list_of_urls[i]})

		// For simplicity, will ignore error for the time being
		_, _ = http.Post(incoming_webhook_url, header_content_type, bytes.NewBuffer(json_value))
	}
}

func SlackAuthenticated(c *gin.Context) {

}

func HandleCommand(c *gin.Context) {

	text := c.PostForm("text")

	response_string := "Hello from the app"

	if "rss" == text {
		guids := GetListOfNewGUIDs()

		if len(guids) == 0 {
			response_string = "Nothing new to show here"
		} else {
			response_string = strings.Join(guids, "\n")
		}
	}

	c.String(200, response_string)

}