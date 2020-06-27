package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/julienrbrt/ut_research_project/recommend"
	"github.com/julienrbrt/ut_research_project/util"
)

//tool arguments
//userID to which user to get recommendations
//nbRecipes is the number of recipes to recommend
//maxDistance define the maximal distance for which users are considered neighbors
func main() {
	//get arguments
	args := os.Args
	if len(args) < 4 {
		fmt.Printf("Error: argument(s) missing, only received %d\nUsage: vinaigrette userID nbRecipes maxDistance\n", len(args))
		os.Exit(1)
	}

	//set user ID
	userID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("Error: userID must be an integer: %v\n", err)
	}

	//number recommendation
	nbRecipes, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Error: nbRecipes must be an integer: %v\n", err)
	}

	//set distance
	maxDistance, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		fmt.Printf("Error: maxDistance must be an integer: %v\n", err)
	}

	//load datasets
	log.Printf("Loading datasets...\n")
	users := util.LoadCSV("data/users.csv")
	orders := util.LoadCSV("data/orders.csv")
	recipes := util.LoadCSV("data/recipes.csv")

	//content filtering
	err = recommend.WithContentFiltering(userID, nbRecipes, users, orders, recipes)
	if err != nil {
		log.Fatalln(err)
	}

	//collaborative filtering
	err = recommend.WithCollaborativeFiltering(userID, nbRecipes, maxDistance, users, orders, recipes)
	if err != nil {
		log.Fatalln(err)
	}
}
