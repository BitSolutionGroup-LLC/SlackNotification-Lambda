package main

// GO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w' -i -o slack-notify main.go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Records []struct {
		SNS struct {
			Type       string `json:"Type"`
			Timestamp  string `json:"Timestamp"`
			SNSMessage string `json:"Message"`
		} `json:"Sns"`
	} `json:"Records"`
}

type SNSMessage struct {
	AlarmName      string `json:"AlarmName"`
	NewStateValue  string `json:"NewStateValue"`
	NewStateReason string `json:"NewStateReason"`
	Channel        string
}

type SlackMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	Channel     string       `json:"channel"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	UserName    string       `json:"username,omitempty"`
}

type Attachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
	Title string `json:"title"`
}

func handler(request Request) error {
	var snsMessage SNSMessage
	snsMessage.Channel = os.Getenv("SLACK_CHANNEL")
	err := json.Unmarshal([]byte(request.Records[0].SNS.SNSMessage), &snsMessage)
	if err != nil {
		value := strings.Split(fmt.Sprintf("%s", request.Records[0].SNS.SNSMessage), ",")
		snsMessage.AlarmName = value[0]
		snsMessage.Channel = value[1]
		snsMessage.NewStateReason = strings.Join(value[2:], ",")
	}

	log.Printf("New alarm: %s - Reason: %s", snsMessage.AlarmName, snsMessage.NewStateReason)
	slackMessage := buildSlackMessage(snsMessage)
	if slackMessage.UserName == "" {
		slackMessage.UserName = "AWS"
	}
	if slackMessage.Channel == "" {
		slackMessage.Channel = "@dicruzg"
	}
	if slackMessage.IconEmoji == "" {
		slackMessage.IconEmoji = ":cloudform:"
	}
	fmt.Printf("%+v\n%+v\n", snsMessage, slackMessage)
	postToSlack(slackMessage)
	log.Println("Notification has been sent")
	return nil
}

func buildSlackMessage(message SNSMessage) SlackMessage {
	return SlackMessage{
		Text:      fmt.Sprintf("`%s`", message.AlarmName),
		Channel:   message.Channel,
		UserName:  os.Getenv("SLACK_USERNAME"),
		IconEmoji: os.Getenv("SLACK_EMOJI"),
		Attachments: []Attachment{
			Attachment{
				Text:  message.NewStateReason,
				Color: "danger",
				Title: "Reason",
			},
		},
	}
}

func postToSlack(message SlackMessage) error {
	client := &http.Client{}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", os.Getenv("SLACK_WEBHOOK"), bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		return err
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
