package main 

import (
    "os"
	"fmt"
	"log"
	"flag"

)

func main() {
	// Define a flag for the command
	message := flag.String("write", "","Command to write file")
	read := flag.String("read","deee","Command to read file")

	// Parse the command-line arguments
	flag.Parse()

	if *message != "" {
		// message := flag.Arg(0)
		fmt.Printf(*message + " ")
		writeFile(*message)
		fmt.Print("Message send")
	} else if *read != ""{
		readFile()
	}
	
}


func writeFile(message string){
	filePath := "data/file.txt"

	//https://gist.github.com/radxene/f5ac02ef1cbb6b91711c534824bd179a
	// Change perrmissions using Linux style
	err := os.Chmod(filePath, 0777)
	if err != nil {
		fmt.Println(err)
	}

	// Create or open the file for writing
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("Error creating the file:", err)
		return
	}
	defer file.Close()

	// Write content to the file
	content := []byte("\n"+message)
	_, err = file.Write(content)
	if err != nil {
		fmt.Println("Error writing to the file:", err)
		return
	}

	fmt.Println("File written successfully.")
}


func readFile() {
	file, err := os.Open("data/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Create a byte slice to hold the data
	data := make([]byte, 1024) // Adjust the size as needed

	// Read data from the file
	n, err := file.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the byte slice to a string for printing
	content := string(data[:n])

	fmt.Println(content)
}







