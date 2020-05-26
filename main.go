package main

import (
	"fmt"
	"log"
)

func main() {
	df, err := LoadDataFromCSV()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(df)
}
