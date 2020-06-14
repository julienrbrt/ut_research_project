package main

import (
	"log"

	"github.com/julienrbrt/ut_research_project/generate"
	"github.com/julienrbrt/ut_research_project/recipe"
)

func main() {
	//Scrape recipes
	recipes, err := recipe.RecipesData(5000, "data/orders.csv")
	if err != nil {
		log.Fatalln(err)
	}
	//Generate user data
	err = generate.UsersData(50000, recipes, "data/users.csv", "data/orders.csv")
	if err != nil {
		log.Fatalln(err)
	}
}
