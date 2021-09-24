package main

import (
	"bytes"
	"context"
	json2 "encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"os"
	"strings"
	"time"
)

type Message struct {
	Text     string
	Username string
}

type Notification struct {
	NotificationType string `json:"notificationType"`
	Bounce           struct {
		BounceType        string `json:"bounceType"`
		ReportingMTA      string `json:"reportingMTA"`
		BouncedRecipients []struct {
			EmailAddress   string `json:"emailAddress"`
			Status         string `json:"status"`
			Action         string `json:"action"`
			DiagnosticCode string `json:"diagnosticCode"`
		} `json:"bouncedRecipients"`
		BounceSubType string    `json:"bounceSubType"`
		Timestamp     time.Time `json:"timestamp"`
		FeedbackID    string    `json:"feedbackId"`
		RemoteMtaIP   string    `json:"remoteMtaIp"`
	} `json:"bounce"`
	Complaint struct {
		UserAgent            string `json:"userAgent"`
		ComplainedRecipients []struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"complainedRecipients"`
		ComplaintFeedbackType string    `json:"complaintFeedbackType"`
		ArrivalDate           time.Time `json:"arrivalDate"`
		Timestamp             time.Time `json:"timestamp"`
		FeedbackID            string    `json:"feedbackId"`
	} `json:"complaint"`
	Mail struct {
		Timestamp        time.Time `json:"timestamp"`
		Source           string    `json:"source"`
		SourceArn        string    `json:"sourceArn"`
		SourceIP         string    `json:"sourceIp"`
		SendingAccountID string    `json:"sendingAccountId"`
		MessageID        string    `json:"messageId"`
		Destination      []string  `json:"destination"`
		HeadersTruncated bool      `json:"headersTruncated"`
		Headers          []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"headers"`
		CommonHeaders struct {
			From      []string `json:"from"`
			Date      string   `json:"date"`
			To        []string `json:"to"`
			MessageID string   `json:"messageId"`
			Subject   string   `json:"subject"`
		} `json:"commonHeaders"`
	} `json:"mail"`
}

func HandleLambdaEvent(ctx context.Context, event events.SNSEvent) {
	client := &http.Client{}
	url := os.Getenv("WEBHOOK_URL")

	for _, record := range event.Records {
		sns := record.SNS

		notification := Notification{}
		err := json2.Unmarshal([]byte(sns.Message), &notification)

		if err != nil {
			panic(err)
		}
		message := Message{}

		message.Text = "# " + notification.NotificationType + " \n" +
			"``` " + " \n" +
			"FROM: " + strings.Join(notification.Mail.CommonHeaders.From, " / ") + " \n" +
			"TO: " + strings.Join(notification.Mail.CommonHeaders.To, " / ") + " \n" +
			"SUBJECT: " + notification.Mail.CommonHeaders.Subject + " \n" +
			"```"

		if notification.NotificationType == "Bounce" {
			message.Text += " \n\n " +
				"### Bounced Recipients: \n"

			for _, r := range notification.Bounce.BouncedRecipients {
				message.Text += "- " + r.EmailAddress + "\n"
			}
		}

		message.Username = "AWS SES"

		json, err := json2.Marshal(message)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(json))

		req.Header.Set("Content-Type", "application/json")

		_, err = client.Do(req)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
