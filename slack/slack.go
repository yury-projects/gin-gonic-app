package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
)

const (
	HEADER_CONTENT_TYPE  = "application/json"
	COMMAND_HELP_MESSAGE = `
	rss - usage: rss - get latest articles from the rss feed
	weather - usage: weather in [city] - get weather forecast for city
	ttc - usage: ttc next (line 1 subway|bus 84) (north|east|south|west) at 'station name'
	hnews - usage: hnews top 50
	help - usage: help - show this message
	`
	UNKNOWN_COMMAND_MESSAGE = ""
)

var (
	INCOMING_WEBHOOK_URL = os.Getenv("GIN_GONIC_SLACK_WEBHOOK_URL")
	VERIFICATION_TOKEN   = os.Getenv("GIN_GONIC_SLACK_VERIFICATION_TOKEN")
	setOfAllowedCommands = map[string]bool{
		"rss":     true,
		"weather": true,
		"ttc":     true,
		"hnews":   true,
		"help":    true,
	}
	listOfAllowedCommands = make([]string, len(setOfAllowedCommands))
	commandRegexp         = "^(?P<command>(rss|weather|ttc|hnews|help))"
	compiledCommandRegexp = regexp.MustCompile(commandRegexp)
)

func init() {
	i := 0

	// Initializing regexp-es
	for commandName := range setOfAllowedCommands {
		listOfAllowedCommands[i] = commandName
		i++
	}

}

func getCommandObj(text string) (interface{}, error) {

	command := compiledCommandRegexp.Find([]byte(text))

	if len(command) == 0 {

		return 0, fmt.Errorf("Unable to find allowed command, type help for more info")
	}

	cmd := string(command)
	fmt.Println(cmd)

	var out interface{}

	switch cmd {
	case "rss":
		out = &RSSCommand{}
		break
	case "weather":
		out = &WeatherCommand{}
		break
	case "ttc":
		out = &TTCCommand{}
		break
	case "hnews":
		out = &HackerNewsCommand{}
		break
	case "help":
		break
	}

	return out, nil

}

func getListOfAllowedCommandsMessage() string {

	return "Unrecognized command. \nCurrently only rss, weather and help commands are allowed."
}

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
		// c.AbortWithStatusJSON(400, "{\"error\": 400, \"msg\":\"Soemthing is wrong with verification token\"")
	}

}

// HandleCommand - general handler function for all verified Slack commands
func HandleCommand(c *gin.Context) {

	text := c.PostForm("text")

	// var responseInterface interface{} = "Hello from the app"

	cmd, err := getCommandObj(text)

	if err != nil {

		c.JSON(200, getListOfAllowedCommandsMessage())
		return
	}

	command := cmd.(CommandInterface)

	// switch command {
	// case "rss":
	// 	guids := GetListOfNewGUIDs()

	// 	if len(guids) == 0 {
	// 		responseInterface = "Nothing new to show here"
	// 	} else {
	// 		responseInterface = strings.Join(guids, "\n")
	// 	}

	// 	cmd = RSSCommand{}
	// case "weather":
	// 	responseInterface = CurrentWeatherFromCity("Toronto")
	// 	cmd = WeatherCommand{}
	// case "ttc":
	// 	cmd = TTCCommand{}
	// case "hnews":
	// 	cmd = HackerNewsCommand{}
	// case "help":
	// 	responseInterface = getListOfAllowedCommandsMessage()
	// default:
	// 	c.JSON(200, getListOfAllowedCommandsMessage())
	// 	return
	// }

	fmt.Println(command.GetCommandResponse())

	c.JSON(200, command.GetCommandResponse())

}
