package main

import (
	"bytes"
	"log"
	"encoding/json"
	"fmt"
	"gopkg.in/gomail.v2"
	"io"
	"math/rand"
	"net/http"
	"text/template"
)

func sendGoMail(templatePath string, Name string) {
	var body bytes.Buffer
	t, err := template.ParseFiles(templatePath)
	t.Execute(&body, struct{ Name string }{Name: Name})

	if err != nil {
		fmt.Println(err)
		return
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "my@gmail.com")
	// m.SetHeader("To", recipientEmail)
	m.SetHeader("To", "your@gmail.com")
	// m.SetAddressHeader("Cc", "cc@gmail.com", "Name")
	m.SetHeader("Subject", "Gita Verse")
	m.SetBody("text/html", body.String())
	m.Attach("./kol.png")

	d := gomail.NewDialer("smtp.gmail.com", 587, "my@gmail.com", "my_email_password")
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}

func main() {

	chapter := rand.Intn(18-1) + 1
	verse := rand.Intn(20-1) + 1
	url := fmt.Sprintf("https://bhagavad-gita3.p.rapidapi.com/v2/chapters/%d/verses/%d/", chapter, verse)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("X-RapidAPI-Key", "your_gita_api_key")
	req.Header.Add("X-RapidAPI-Host", "bhagavad-gita3.p.rapidapi.com")

	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode >= 400 {
		log.Println("Request failed with status:", res.StatusCode)
		return // Or handle the error as needed
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	// fmt.Println(res)
	// fmt.Println(string(body))
	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Printf("could not unmarshal json: %s\n", err)
		return
	}

	// fmt.Printf("json map: %v\n", data)
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
	// file, err := os.Open("./mail.txt")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer file.Close()

	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	email := scanner.Text()
	// 	sendGoMail(string(email), "./test.html", text)
	// }
	sendGoMail("./test.html", text)
}
