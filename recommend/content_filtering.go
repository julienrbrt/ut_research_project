package recommend

import (
	"math"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

//UserProfile returns a dataframe containing orders and their normalized rating
func UserProfile(userID int, orders, recipes dataframe.DataFrame) dataframe.DataFrame {
	//filter the matching user
	normalizedOrders := orders.
		Filter(dataframe.F{Colname: "user_id", Comparator: series.Eq, Comparando: userID})

	//normalize ratings
	ratingsAvg := normalizedOrders.Col("rating").Map(func(e series.Element) series.Element {
		result := e.Copy()
		result.Set(result.Float() - normalizedOrders.Col("rating").Mean())
		return series.Element(result)
	})

	//mutate dataframe by changing rating and inner join with recipes dataframe
	normalizedOrders = normalizedOrders.
		Mutate(ratingsAvg).
		InnerJoin(recipes.Copy().Rename("recipe_id", "id"), "recipe_id")

	//keep only relevant columns
	var columnsToDrop []string
	for _, n := range normalizedOrders.Names() {
		//keep only orders columns and recipes tags
		if n != "user_id" && n != "recipe_id" && n != "rating" && !strings.Contains(n, "tag_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}

	normalizedOrders = normalizedOrders.Drop(columnsToDrop)

	return normalizedOrders
}

//UserTagsWeight will use the UserProfile output to generate the recipes tags weight based on the ratings of each recipes using that tag
func UserTagsWeight(orders dataframe.DataFrame) map[string]float64 {
	weight := make(map[string]float64)
	for _, n := range orders.Names() {
		if strings.Contains(n, "tag_") {
			//calculate the weight
			w := orders.
				Filter(dataframe.F{Colname: n, Comparator: series.Eq, Comparando: "1"}).
				Col("rating").
				Mean()
				//only add tags weight with value
			if !math.IsNaN(w) {
				weight[n] = w
			}
		}
	}

	//workaround to create a dataframe from only one row
	return weight
}

//WithContentFiltering recommends recipes using content filtering
func WithContentFiltering() {

}

///content filtering

//create user profile (normalized ratings with recipes tags)
// orders = recommend.UserProfile(userID, orders, recipes)
// fmt.Printf("User %d has made %d orders with a (normalized) average rating of %.2f per order\n", userID, orders.Nrow(), orders.Col("rating").Mean())

// _ = recommend.UserTagsWeight(orders)
