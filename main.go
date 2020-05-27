package main

import (
	"fmt"
	"log"
)

func main() {
	// df, err := LoadData(4950, true)
	df, err := LoadDataFromCSV()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(df)
}
