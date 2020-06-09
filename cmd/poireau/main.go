package main

import (
	"fmt"
	"path"

	"github.com/kardianos/osext"

	"github.com/julienrbrt/ut_research_project/generate"
	"github.com/julienrbrt/ut_research_project/recipe"
	"github.com/julienrbrt/ut_research_project/util"
)

func main() {
	//Get Executable Flolder
	ef, err := osext.ExecutableFolder()
	if err != nil {
		panic(err)
	}

	//Load datasets
	users := util.LoadCSV(path.Join(ef, generate.UsersCSVPath))
	orders := util.LoadCSV(path.Join(ef, generate.OrdersCSVPath))
	recipes := util.LoadCSV(path.Join(ef, recipe.RecipesCSVPath))

	fmt.Println(users)
	fmt.Println(orders)
	fmt.Println(recipes)
}
