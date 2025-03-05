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

func main() {
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
	sender := "muterjeffery@gmail.com"    // Must be verified in SES
	recipient := "muterjeffery@gmail.com" // Must be verified if in sandbox mode
	subject := "Hello from AWS SES (Go)"
	textBody := "This is a test email sent using AWS SES from Go."
	htmlBody := "<h1>Hello!</h1><p>This is a test email sent using <strong>AWS SES</strong> from Go.</p>"

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
