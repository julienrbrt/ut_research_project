package main

import "log"

func main() {
	recipes := scrapeXAH(10)
	header, records := recipes.transformCSV()
	err := writeCSV("data/item_recipes.csv", header, records)
	if err != nil {
		log.Fatalln(err)
	}
}
