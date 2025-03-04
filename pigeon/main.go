package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("begin sending email...")

	// Gmail SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Your Gmail credentials

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Get credentials from .env
	fromEmail := os.Getenv("GMAIL_EMAIL")
	password := os.Getenv("GMAIL_PASSWORD")

	// Recipient email address
	toEmail := []string{"recipient@example.com"}

	// Email content
	subject := "Test Email from Go"
	body := `
	<html>
		<body>
			<h1>Hello from Go!</h1>
			<p>This is a <b>formatted</b> email sent using Go.</p>
			<p>Here's some technical information:</p>
			<pre>
			func sendEmail() {
				// This is a code example
				fmt.Println("Sending email...")
			}
			</pre>
			<p>Thank you!</p>
		</body>
	</html>
	`

	// Build the email
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n%s\r\n%s",
		strings.Join(toEmail, ","),
		subject,
		mime,
		body))

	// Authentication
	auth := smtp.PlainAuth("", fromEmail, password, smtpHost)

	// Send email
	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, fromEmail, toEmail, message)
	if err != nil {
		fmt.Printf("Error sending email: %s", err)
		return
	}

	fmt.Println("Email sent successfully!")
}
