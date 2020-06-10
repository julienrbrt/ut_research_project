package recommend

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

//UserProfile creates an user profile dataframe with normalized ratings
func UserProfile(userID int, orders dataframe.DataFrame) dataframe.DataFrame {
	ratingsUser := orders.
		Filter(dataframe.F{Colname: "user_id", Comparator: series.Eq, Comparando: userID})

	ratingsAvg := ratingsUser.Col("rating").Map(func(e series.Element) series.Element {
		result := e.Copy()
		result.Set(result.Float() - ratingsUser.Col("rating").Mean())
		return series.Element(result)
	})

	return ratingsUser.Mutate(ratingsAvg)
}
