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
//maxDistance define the maximal distance for which users are considered neighbors
func main() {
	//get arguments
	args := os.Args
	if len(args) < 3 {
		fmt.Printf("Error: argument(s) missing, only received %d\nUsage: vinaigrette userID maxDistance\n", len(args))
		os.Exit(1)
	}

	//set user ID
	userID, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("Error: userID must be an integer: %v\n", err)
	}

	//set distance
	maxDistance, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		fmt.Printf("Error: maxDistance must be an integer: %v\n", err)
	}

	//Load datasets
	log.Println("Loading datasets...")
	users := util.LoadCSV("data/users.csv")
	orders := util.LoadCSV("data/orders.csv")
	recipes := util.LoadCSV("data/recipes.csv")

	///collaborative filtering

	//filter by neighbors users
	users = recommend.UsersCloseByXKm(userID, maxDistance, users)

	///content filtering

	//create user profile
	_ = recommend.UserProfile(userID, orders)

	//build users liked tags
	recommend.UserRecipesTags(userID, orders, recipes)
}
