package main

import (
	"fmt"
	"log"
)

func main() {
	df, err := LoadData(4000, true)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(df)
}
