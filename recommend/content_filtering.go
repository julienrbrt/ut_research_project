package recommend

import (
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
func UserRecipesTags(userID int, orders, recipes dataframe.DataFrame) (dataframe.DataFrame, error) {
	//filter the matching user
	ratingsUser := orders.
		Filter(dataframe.F{Colname: "user_id", Comparator: series.Eq, Comparando: userID})

	//get list or user orders
	orderID, err := ratingsUser.Col("order_id").Int()
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	//filter recipes
	var filters []dataframe.F
	for _, o := range orderID {
		filters = append(filters, dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: o})
	}

	recipes.
		FilterAggregation(
			dataframe.Or,
			filters...)

	return dataframe.DataFrame{}, nil
}
