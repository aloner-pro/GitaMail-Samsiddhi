package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"text/template"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

// Idata represents the structure of the JSON data from the API response
type Idata struct {
	Id    float32 `json:"id"`
	Vn    float32 `json:"verse_number"`
	Cp    float32 `json:"chapter_number"`
	Sl    string  `json:"slug"`
	Text  string  `json:"text"`
	Ts    string  `json:"transliteration"`
	Wm    string  `json:"word_meanings"`
	Trans []struct {
		Id   float32 `json:"id"`
		Des  string  `json:"description"`
		Auth string  `json:"author_name"`
		Lang string  `json:"language"`
	} `json:"translations"`
	Coms []struct {
		Id   float32 `json:"id"`
		Desp string  `json:"description"`
		Au   string  `json:"author_name"`
		Lg   string  `json:"language"`
	} `json:"commentaries"`
}

// sendGoMail sends an email using the provided template and data to a list of recipients
func sendGoMail(templatePath string, data interface{}, to []string) error {
	// Parse the email template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	// Load the configuration from the YAML file
	vi := viper.New()
	vi.SetConfigFile("test.yaml")
	err = vi.ReadInConfig()
	if err != nil {
		return err
	}

	// Create a new SMTP dialer
	d := gomail.NewDialer("smtp.gmail.com", 587, vi.GetString("mymail"), vi.GetString("mymailpassword"))

	// Send email to each recipient individually
	for _, recipient := range to {
		var body bytes.Buffer
		err := t.Execute(&body, data)
		if err != nil {
			return err
		}

		m := gomail.NewMessage()
		m.SetHeader("From", vi.GetString("mymail"))
		m.SetHeader("To", recipient)
		m.SetHeader("Subject", "Gita Verse")
		m.SetBody("text/html", body.String())

		// Dial and send the email
		err = d.DialAndSend(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// Generate random chapter and verse numbers
	chapter := rand.Intn(18-1) + 1
	verse := rand.Intn(20-1) + 1
	url := fmt.Sprintf("https://bhagavad-gita3.p.rapidapi.com/v2/chapters/%d/verses/%d/", chapter, verse)

	// Create a new HTTP request
	req, _ := http.NewRequest("GET", url, nil)

	// Load the configuration from the YAML file
	vi := viper.New()
	vi.SetConfigFile("test.yaml")
	vi.ReadInConfig()

	// Add API key and host to the request headers
	req.Header.Add("X-RapidAPI-Key", vi.GetString("myapikey"))
	req.Header.Add("X-RapidAPI-Host", "bhagavad-gita3.p.rapidapi.com")

	// Send the HTTP request and get the response
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode >= 400 {
		log.Println("Request failed with status:", res.StatusCode)
		return // Or handle the error as needed
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// Parse the JSON response into the Idata struct
	var idata Idata
	err := json.Unmarshal(body, &idata)
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return
	}

	// Print the raw text of the verse
	rawText := idata.Text
	fmt.Printf("%s\n", rawText)

	// Extract the English translation if available
	var trans string
	if len(idata.Trans) > 0 && idata.Trans[0].Lang == "english" {
		trans = idata.Trans[0].Des
		fmt.Printf("Meaning:\n %+v\n", trans)
	}

	// Prepare data for email template
	data := struct {
		Verse   string
		Meaning string
	}{
		Verse:   idata.Text,
		Meaning: trans,
	}

	// Create a new client for interacting with Airtable
	client, _ := New(vi.GetString("airtabletoken"), vi.GetString("baseid"))

	// Define the structure for Airtable records
	type task struct {
		AirtableID string
		Fields     struct {
			Name       string
			Email      string
			Interested string
			Date       string
		}
	}

	// Fetch the records from Airtable
	tasks := []task{}
	if err := client.ListRecords("mail", &tasks); err != nil {
		panic(err)
	}

	// Collect emails of interested recipients
	var emails []string
	for _, t := range tasks {
		if t.Fields.Interested == "Yes" {
			emails = append(emails, t.Fields.Email)
		}
	}

	fmt.Println("All emails:", emails)

	// Send the email to all interested recipients
	err = sendGoMail("./template.html", data, emails)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Emails sent successfully!")
}
