package main

import (
	"fmt"
	"bufio"
	"os"
)

func main() {

	// Open the file containing email addresses
	emailsFile, err := os.Open("emails.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	// Ensure the file is closed when the function exits
	defer emailsFile.Close()

	// Slice to store the email addresses
	to := []string{}
	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(emailsFile)
	for scanner.Scan() {
		to = append(to, scanner.Text())
	}

	// Check for any errors encountered while reading the file
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Println("All emails:", to)
}
