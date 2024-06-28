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

func sendGoMail(templatePath string, data interface{}, to []string) error {
	// Parse the email template
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	// Load the configuration
	vi := viper.New()
	vi.SetConfigFile("test.yaml")
	err = vi.ReadInConfig()
	if err != nil {
		return err
	}

	// Create a new dialer
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

		err = d.DialAndSend(m)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {

	chapter := rand.Intn(18-1) + 1
	verse := rand.Intn(20-1) + 1
	url := fmt.Sprintf("https://bhagavad-gita3.p.rapidapi.com/v2/chapters/%d/verses/%d/", chapter, verse)

	req, _ := http.NewRequest("GET", url, nil)

	vi := viper.New()
	vi.SetConfigFile("test.yaml")
	vi.ReadInConfig()

	req.Header.Add("X-RapidAPI-Key", vi.GetString("myapikey"))
	req.Header.Add("X-RapidAPI-Host", "bhagavad-gita3.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode >= 400 {
		log.Println("Request failed with status:", res.StatusCode)
		return // Or handle the error as needed
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)


	var idata Idata
	err := json.Unmarshal(body, &idata)
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return
	}

	rawText := idata.Text
	fmt.Printf("%s\n", rawText)
	var trans string
	if idata.Trans[0].Lang == "english" {
		trans = idata.Trans[0].Des
		fmt.Printf("Meaning:\n %+v\n", trans)
	}

	data := struct {
		Verse    string
		Meaning string
	}{
		Verse:   idata.Text,
		Meaning: trans,
	}

	client, _ := New(vi.GetString("airtabletoken"), vi.GetString("baseid"))

	type task struct {
		AirtableID string
		Fields     struct {
			Name  string
			Email string
			Interested string
			Date  string
		}
	}
	tasks := []task{}
	if err := client.ListRecords("mail", &tasks); err != nil {
		panic(err)
	}

	var emails []string
	for _, t := range tasks {
		if t.Fields.Interested == "Yes" {
			emails = append(emails, t.Fields.Email)
		}
	}

	fmt.Println("All emails:", emails)

	err = sendGoMail("./template.html", data, emails)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Emails sent successfully!")
}
