package main

import "log"

func main() {
	//Scrape
	recipes := ScrapeXAH(10)
	header, records := recipes.TransformToCSV()
	// err := WriteCSV("data/item_recipes.csv", header, records)

	//Clean
	err := CleanCSV(header, records)
	if err != nil {
		log.Fatalln(err)
	}

	//Generate

	//Recommend
}
