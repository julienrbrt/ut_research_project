package main

import (
	"fmt"

	"github.com/julienrbrt/ut_research_project/recommend"
	"github.com/julienrbrt/ut_research_project/util"
)

func main() {
	//Load datasets
	// users := util.LoadCSV("data/users.csv")
	orders := util.LoadCSV("data/orders.csv")
	// recipes := util.LoadCSV("data/recipes.csv")
	fmt.Println(recommend.UserProfile(5, orders))
}
