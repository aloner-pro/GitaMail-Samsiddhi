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

func sendGoMail(templatePath string, Text string, to []string) error {
	var body bytes.Buffer
	t, err := template.ParseFiles(templatePath)
	t.Execute(&body, struct{ Text string }{Text: Text})

	if err != nil {
		fmt.Println(err)
		return nil
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


	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return
	}

	rawText, ok := data["text"]
	if !ok {
		fmt.Printf("text does not exist\n")
		return
	}
	text, ok := rawText.(string)
	if !ok {
		fmt.Printf("text is not a string\n")
		return
	}
	fmt.Printf("%s\n", text)
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
	err = sendGoMail("./OKTEST.html", text, to)
	if err != nil {
		fmt.Println("Error sending email:", err)
		return
	}
	fmt.Println("Emails sent successfully!")
}
