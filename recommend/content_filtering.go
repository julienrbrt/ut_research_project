package recommend

import (
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

//UserProfile creates an user profile dataframe with normalized ratings
func UserProfile(userID int, orders dataframe.DataFrame) dataframe.DataFrame {
	//filter the matching user
	ratingsUser := orders.
		Filter(dataframe.F{Colname: "user_id", Comparator: series.Eq, Comparando: userID})

	//normalize ratings
	ratingsAvg := ratingsUser.Col("rating").Map(func(e series.Element) series.Element {
		result := e.Copy()
		result.Set(result.Float() - ratingsUser.Col("rating").Mean())
		return series.Element(result)
	})

	return ratingsUser.Mutate(ratingsAvg)
}

//UserRecipesTags returns a dataframe containing a list of tags present in the users rated recipes
func UserRecipesTags(userID int, orders, recipes dataframe.DataFrame) dataframe.DataFrame {
	//filter the matching user
	ratingsUser := orders.
		Filter(dataframe.F{Colname: "user_id", Comparator: series.Eq, Comparando: userID})

	//get list of user orders
	ordersUser := ratingsUser.InnerJoin(recipes.Copy().Rename("order_id", "id"), "order_id")

	//keep only relevant columns
	var columnsToDrop []string
	for _, n := range ordersUser.Names() {
		//keep only tags
		if n != "order_id" && n != "rating" && !strings.Contains(n, "tag_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}

	ordersUser = ordersUser.Drop(columnsToDrop)

	return ordersUser
}
