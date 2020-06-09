package main

import (
	"log"

	"github.com/julienrbrt/ut_research_project/generate"
	"github.com/julienrbrt/ut_research_project/recipe"
)

func main() {
	//Scrape recipes
	recipes, err := recipe.RecipesData(5000, true)
	if err != nil {
		log.Fatalln(err)
	}
	//Generate user data
	err = generate.UsersData(50000, recipes, true)
	if err != nil {
		log.Fatalln(err)
	}
}
