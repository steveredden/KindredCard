package mailer

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
	"github.com/steveredden/KindredCard/internal/models"
	"github.com/steveredden/KindredCard/internal/utils"
)

type Config struct {
	Host string
	Port string
	User string
	Pass string
	From string
}

type EmailContent struct {
	Subject string
	Body    string
}

func LoadConfig() Config {
	return Config{
		Host: os.Getenv("SMTP_HOST"),
		Port: os.Getenv("SMTP_PORT"),
		User: os.Getenv("SMTP_USER"),
		Pass: os.Getenv("SMTP_PASS"),
		From: os.Getenv("SMTP_FROM"),
	}
}

func SendEventNotification(to, subject, body string) error {
	c := LoadConfig()
	if c.Host == "" || c.Port == "" || c.User == "" || c.Pass == "" || c.From == "" {
		logger.Error("[MAILER] Missing required SMTP environment variables!")
		return nil
	}

	// Standard HTML Email Headers
	header := make(map[string]string)
	header["From"] = c.From
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""

	var message string
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	auth := smtp.PlainAuth("", c.User, c.Pass, c.Host)
	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)

	return smtp.SendMail(addr, auth, c.User, []string{to}, []byte(message))
}

// BuildTodayEventsBody creates an HTML body mimicking a Discord embed
func BuildTodayEventsBody(events []models.UpcomingEvent, baseURL string) EmailContent {
	// Template for the "Embed" style email
	const emailTemplate = `
	<div style="font-family: sans-serif; max-width: 600px; border-left: 4px solid #5865F2; padding: 20px; background-color: #f9f9f9; border-radius: 4px;">
		<h2 style="color: #1a1a1a; margin-top: 0;">{{.Title}}</h2>
		
		{{if not .Events}}
			<p style="color: #4a4a4a;">No birthdays or anniversaries today!</p>
		{{else}}
			{{if .Birthdays}}
				<h3 style="color: #5865F2; margin-bottom: 5px;">Birthdays</h3>
				<ul style="list-style: none; padding-left: 0;">
					{{range .Birthdays}}
					<li style="margin-bottom: 8px;">🎂 <strong><a href="{{$.BaseURL}}/contacts/{{.ContactID}}" style="color: #5865F2; text-decoration: none;">{{.FullName}}</a></strong> - {{.Description}}</li>
					{{end}}
				</ul>
			{{end}}

			{{if .Anniversaries}}
				<h3 style="color: #5865F2; margin-bottom: 5px;">Anniversaries</h3>
				<ul style="list-style: none; padding-left: 0;">
					{{range .Anniversaries}}
					<li style="margin-bottom: 8px;">💍 <strong><a href="{{$.BaseURL}}/contacts/{{.ContactID}}" style="color: #5865F2; text-decoration: none;">{{.FullName}}</a></strong> - {{.Description}}</li>
					{{end}}
				</ul>
			{{end}}

			{{if .Others}}
				<h3 style="color: #5865F2; margin-bottom: 5px;">Other Dates</h3>
				<ul style="list-style: none; padding-left: 0;">
					{{range .Others}}
					<li style="margin-bottom: 8px;">📅 <strong><a href="{{$.BaseURL}}/contacts/{{.ContactID}}" style="color: #5865F2; text-decoration: none;">{{.FullName}}</a></strong> - {{.Description}}</li>
					{{end}}
				</ul>
			{{end}}
		{{end}}

		<hr style="border: 0; border-top: 1px solid #e0e0e0; margin: 20px 0;">
		<p style="font-size: 12px; color: #7a7a7a;">Sent by KindredCard</p>
	</div>`

	// Grouping logic
	data := struct {
		Title         string
		BaseURL       string
		Events        []models.UpcomingEvent
		Birthdays     []map[string]interface{}
		Anniversaries []map[string]interface{}
		Others        []map[string]interface{}
	}{
		Title:   fmt.Sprintf("🎉 Today's Events - %s", time.Now().Local().Format("Jan 2")),
		BaseURL: baseURL,
		Events:  events,
	}

	for _, e := range events {
		desc := ""
		if e.AgeOrYears != nil {
			ordinal := utils.Ordinal(*e.AgeOrYears)
			switch e.EventType {
			case "birthday":
				desc = fmt.Sprintf("%s birthday %s!", ordinal, e.TimeDescription)
			case "anniversary":
				desc = fmt.Sprintf("%s wedding anniversary %s!", ordinal, e.TimeDescription)
			default:
				desc = fmt.Sprintf("%s anniversary of %s %s!", ordinal, e.EventType, e.TimeDescription)
			}
		} else {
			switch e.EventType {
			case "birthday":
				desc = fmt.Sprintf("has a birthday %s!", e.TimeDescription)
			case "anniversary":
				desc = fmt.Sprintf("has a wedding anniversary %s!", e.TimeDescription)
			default:
				desc = fmt.Sprintf("anniversary of %s %s!", e.EventType, e.TimeDescription)
			}
		}

		item := map[string]interface{}{
			"FullName":    e.FullName,
			"ContactID":   e.ContactID,
			"Description": desc,
		}

		switch e.EventType {
		case "birthday":
			data.Birthdays = append(data.Birthdays, item)
		case "anniversary":
			data.Anniversaries = append(data.Anniversaries, item)
		default:
			data.Others = append(data.Others, item)
		}
	}

	tmpl, _ := template.New("email").Parse(emailTemplate)
	var out bytes.Buffer
	tmpl.Execute(&out, data)

	return EmailContent{
		Subject: "KindredCard Event Summary",
		Body:    out.String(),
	}
}

// SendTestNotification sends a test notification with dummy data
func SendTestNotification(recipient string, baseURL string) error {
	dummyAge1 := 30
	dummyAge2 := 2

	dummyEvents := []models.UpcomingEvent{
		{
			ContactID:       99999,
			FullName:        "John Doe",
			EventType:       "birthday",
			AgeOrYears:      &dummyAge1,
			TimeDescription: "Today",
		},
		{
			ContactID:       99998,
			FullName:        "Jane Smith",
			EventType:       "anniversary",
			AgeOrYears:      &dummyAge2,
			TimeDescription: "Tomorrow",
		},
		{
			ContactID:       99997,
			FullName:        "Jack Jones",
			EventType:       "Retirement",
			AgeOrYears:      &dummyAge2,
			TimeDescription: "in 3 days",
		},
	}

	body := BuildTodayEventsBody(dummyEvents, baseURL)
	return SendEventNotification(recipient, "KindredCard Event Summary", body.Body)
}
