package slack

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yury-projects/gin-gonic-app/rss"
	"github.com/yury-projects/gin-gonic-app/weather"
)

const HEADER_CONTENT_TYPE = "application/json"

var (
	INCOMING_WEBHOOK_URL = os.Getenv("GIN_GONIC_SLACK_WEBHOOK_URL")
	VERIFICATION_TOKEN   = os.Getenv("GIN_GONIC_SLACK_VERIFICATION_TOKEN")
)

func NotifyNewContent(list_of_urls []string) {

	for i := len(list_of_urls) - 1; i >= 0; i-- {

		jsonString, _ := json.Marshal(map[string]string{"text": list_of_urls[i]})

		// For simplicity, will ignore error for the time being
		_, _ = http.Post(INCOMING_WEBHOOK_URL, HEADER_CONTENT_TYPE, bytes.NewBuffer(jsonString))
	}
}

func SlackAuthenticated(c *gin.Context) {

}

// AuthenticateCommand - middleware function used to authenticate/verify incoming Slack commands
func AuthenticateCommand(c *gin.Context) {

	verificationToken := c.PostForm("token")

	if verificationToken != VERIFICATION_TOKEN {
		c.AbortWithStatusJSON(400, "{\"error\": 400, \"msg\":\"Soemthing is wrong with verification token\"")
	}

}

// HandleCommand - general handler function for all verified Slack commands
func HandleCommand(c *gin.Context) {

	text := c.PostForm("text")

	var responseInterface interface{} = "Hello from the app"

	if "rss" == text {
		guids := rss.GetListOfNewGUIDs()

		if len(guids) == 0 {
			responseInterface = "Nothing new to show here"
		} else {
			responseInterface = strings.Join(guids, "\n")
		}
	} else if "weather" == text {

		responseInterface = weather.CurrentWeatherFromCity("Toronto")

	}

	c.JSON(200, responseInterface)

}
