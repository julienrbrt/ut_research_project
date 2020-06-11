package recommend

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
)

//UsersCloseByXKm returns a dataframe containings users around user with userID from x kms
func UsersCloseByXKm(usersID int, km float64, users dataframe.DataFrame) dataframe.DataFrame {
	//get user
	user := users.Filter(dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: usersID})
	if user.Nrow() == 0 {
		return dataframe.DataFrame{}
	}

	//get user's position bounding coordinates
	boundingCoordinates := util.BoundingCoordinates(user.Col("latitude").Elem(0).Float(), user.Col("longitude").Elem(0).Float(), km)

	matchingUsers := users.
		FilterAggregation(
			dataframe.And,
			dataframe.F{Colname: "latitude", Comparator: series.GreaterEq, Comparando: boundingCoordinates[0].Latitude},
			dataframe.F{Colname: "longitude", Comparator: series.GreaterEq, Comparando: boundingCoordinates[0].Longitude},
			dataframe.F{Colname: "latitude", Comparator: series.LessEq, Comparando: boundingCoordinates[1].Latitude},
			dataframe.F{Colname: "longitude", Comparator: series.LessEq, Comparando: boundingCoordinates[1].Longitude},
		)

	return matchingUsers
}
