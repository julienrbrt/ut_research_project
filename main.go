package main

import (
	"fmt"
	"log"
)

func main() {
	df, err := LoadData(5000, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(df)
}
