package main

func main() {
	//Scrape and Process
	// recipes := LoadRecipesData(4950, true)
	recipes := LoadCSV(recipesCSVPath)

	//Generate
	users, orders := LoadUsersData(50000, recipes, true)
	// users := LoadCSV(usersCSVPath)
	// orders := LoadCSV(ordersCSVPath)

	//Recommend
	BPRRecommender(orders, users, recipes)
}
