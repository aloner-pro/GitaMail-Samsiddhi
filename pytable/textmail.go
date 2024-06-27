package main

import (
	"fmt"
	"bufio"
	"os"
)

func main() {

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
	fmt.Println("All emails:", to)
}

