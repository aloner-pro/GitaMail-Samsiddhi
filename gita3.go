package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
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
	var body bytes.Buffer
	t, err := template.ParseFiles(templatePath)
	// t.Execute(&body, struct{ Text string }{Text: Text})
	err = t.Execute(&body, data)
	if err != nil {
		return err
	}

	vi := viper.New()
	vi.SetConfigFile("test.yaml")
	vi.ReadInConfig()

	m := gomail.NewMessage()
	m.SetHeader("From", vi.GetString("mymail"))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", "Gita Verse")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer("smtp.gmail.com", 587, vi.GetString("mymail"), vi.GetString("mymailpassword"))
	return d.DialAndSend(m)
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

	emailsFile, err := os.Open("emails.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer emailsFile.Close()
	to := []string{}
	scanner := bufio.NewScanner(emailsFile)
	for scanner.Scan() {
		to = append(to, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	err = sendGoMail("./OKTEST.html", data, to)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Emails sent successfully!")
}
