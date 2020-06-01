package main

func main() {
	//Scrape and Process
	// recipes := LoadRecipesData(4950, true)
	recipes := LoadCSV(recipesCSVPath)

	//Generate
	_, _ = LoadUsersData(5, recipes, true)
	// users := LoadCSV(usersCSVPath)
	// orders := LoadCSV(ordersCSVPath)

	//Recommend
}
