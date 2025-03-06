package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/joho/godotenv"
)

type alertData struct {
	alertTitle          string
	deviceType          string
	deviceName          string
	eventTime           string
	eventDate           string
	deviceLocation      string
	conversationSummary string
	alertQuote          string
	fileUrl             string
}

func main() {
	var alertData alertData

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get credentials from .env
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	// 1. Configure AWS session
	// Replace with your AWS region and credentials
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// 2. Create an SES service client
	svc := ses.New(sess)

	// 3. Prepare email content
	recipient := "jefferymuter@yahoo.com" // Must be verified in SES
	sender := "muterjeffery@gmail.com"    // Must be verified if in sandbox mode
	subject := "Hello from AWS SES (Go)"
	textBody := "This is a test email sent using AWS SES from Go."

	htmlBody := `<h1>Muse: ` + alertData.alertTitle + ` Detected</h1>
		<br>
		<p>Muse has discovered an ` + alertData.alertTitle + ` from: ` + alertData.deviceType + ` ` + alertData.deviceName + `</p>
		<br>
		<h3>Alert triggered by quote:</h3>
		<br> 
		` + alertData.alertQuote + `
		<br>
		<h3>Conversation Summary:</h3>
		<br>
		` + alertData.conversationSummary + `
		<br>
		<h2>Estimated Event Details:</h3>
		<br>
		<h3>Time: ~` + alertData.eventTime + `</h3>
		<br>
		<h3>Date: ` + alertData.eventDate + `</h3>
		<br>
		<h3>Link to file ` + alertData.fileUrl + `<br>`

	// 4. Specify email parameters
	input := &ses.SendEmailInput{
		Source: aws.String(sender),
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBody),
				},
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBody),
				},
			},
		},
	}

	// 5. Send the email
	result, err := svc.SendEmail(input)
	if err != nil {
		log.Fatalf("Error sending email: %v", err)
	}

	fmt.Printf("Email sent successfully! Message ID: %s\n", *result.MessageId)
}
