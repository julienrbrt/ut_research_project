package recommend

import (
	"log"
	"strconv"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
	"github.com/zhenghaoz/gorse/core"
)

//usersCloseByXKm returns a dataframe containings users around user with userID from x kms
func usersCloseByXKm(userID int, km float64, users dataframe.DataFrame) dataframe.DataFrame {
	//get user
	user := users.Filter(dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: userID})
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

//MeasureCollaborativeSellability measures the sellability using the cosine similarity of target users recommendation to neighboring users recommendation
func MeasureCollaborativeSellability(userID, nbRecipes int, km float64, recommendations []string, m core.ModelInterface, data *core.DataSet, train, test core.DataSetInterface, users dataframe.DataFrame) float64 {
	if users.Nrow() == 0 {
		return 1
	}

	//convert recommendation to []float64
	recommendationsFloat := make([]float64, len(recommendations))
	for i, r := range recommendations {
		if n, err := strconv.ParseFloat(r, 64); err == nil {
			recommendationsFloat[i] = n
		}
	}

	sellability := 0.0
	for _, id := range users.Col("id").Records() {
		//fit model
		m.Fit(train, nil)
		//generate recommendations for user
		//get all items in the full dataset
		items := core.Items(data)
		//get user ratings in the training set
		excludeItems := train.User(id)
		//get top recommended items (excluding rated items)
		recommendItems, _ := core.Top(items, id, nbRecipes, excludeItems, m)

		//convert recommendItems to []float64
		recommendItemsFloat := make([]float64, len(recommendItems))
		for i, r := range recommendItems {
			if n, err := strconv.ParseFloat(r, 64); err == nil {
				recommendItemsFloat[i] = n
			}
		}

		sim, err := util.CosineSimilarity(recommendationsFloat, recommendItemsFloat)
		if err != nil {
			continue
		}

		sellability = sellability + sim
	}

	//mean cosine similarity of all users
	sellability = sellability / float64(users.Nrow())

	log.Println(sellability)

	return sellability
}

//MeasureContentSellability measures the sellability using the cosine similarity of target users recommendation to neighboring users recommendation
func MeasureContentSellability(userID int, km float64, recommendations []string) float64 {

	return 0
}
