package recommend

import (
	"fmt"
	"strconv"
	"strings"

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
func MeasureCollaborativeSellability(userID, nbRecipes int, km float64, recommendations []string, m core.ModelInterface, data *core.DataSet, train, test core.DataSetInterface, users, recipes dataframe.DataFrame) float64 {
	if users.Nrow() == 0 {
		return 1
	}

	//keep only relevant recipes columns
	var columnsToDrop []string
	for _, n := range recipes.Names() {
		//keep only orders columns and recipes tags
		if n != "id" && !strings.Contains(n, "tag_") && !strings.Contains(n, "ingredient_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}
	recipes = recipes.Drop(columnsToDrop)

	//get item profile of recommendation
	var err error
	recommendationsProfile := make([][]float64, len(recommendations))
	for i, r := range recommendations {
		recommendationsProfile[i], err = util.SS2SF(recipes.Filter(dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: r}).Records()[1][1:])
		if err != nil {
			continue
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

		//get item profile of neighbors recommendation
		neighborsRecommendationsProfile := make([][]float64, len(recommendItems))
		for i, r := range recommendItems {
			neighborsRecommendationsProfile[i], err = util.SS2SF(recipes.Filter(dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: r}).Records()[0][1:])
			if err != nil {
				continue
			}
		}

		//keep max cosine similarity of a recipe
		var sim, newSim, meanSim float64
		for i := range recommendationsProfile {
			for j := range neighborsRecommendationsProfile {
				newSim, err = util.CosineSimilarity(recommendationsProfile[i], neighborsRecommendationsProfile[j])
				if err != nil {
					continue
				}

				if newSim > sim {
					sim = newSim
				}
			}

			meanSim = +sim
		}

		sellability = sellability + (meanSim / float64(len(recommendationsProfile)))
	}

	//mean cosine similarity of all users
	sellability = sellability / float64(users.Nrow())

	return sellability
}

//MeasureContentSellability measures the sellability using the cosine similarity of target users recommendation to neighboring users recommendation
func MeasureContentSellability(userID, nbRecipes int, km float64, recommendations []string, users, orders, recipes dataframe.DataFrame) float64 {
	if users.Nrow() == 0 {
		return 1
	}

	//keep only relevant recipes columns
	var columnsToDrop []string
	for _, n := range recipes.Names() {
		//keep only orders columns and recipes tags
		if n != "id" && !strings.Contains(n, "tag_") && !strings.Contains(n, "ingredient_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}
	recipes = recipes.Drop(columnsToDrop)

	//get item profile of recommendation
	var err error
	recommendationsProfile := make([][]float64, len(recommendations))
	for i, r := range recommendations {
		recommendationsProfile[i], err = util.SS2SF(recipes.Filter(dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: r}).Records()[1][1:])
		if err != nil {
			continue
		}
	}

	sellability := 0.0
	for _, id := range users.Col("id").Records() {
		sid, _ := strconv.Atoi(id)
		recommendItems, err := recommendedContentFiltering(sid, nbRecipes, km, users, orders, recipes)
		if err != nil {
			continue
		}

		//get item profile of neighbors recommendation
		neighborsRecommendationsProfile := make([][]float64, len(recommendItems))
		for i, r := range recommendItems {
			neighborsRecommendationsProfile[i], err = util.SS2SF(recipes.Filter(dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: r}).Records()[1][1:])
			if err != nil {
				continue
			}
		}

		//keep max cosine similarity of a recipe
		var sim, newSim, meanSim float64
		for i := range recommendationsProfile {
			for j := range neighborsRecommendationsProfile {
				newSim, err = util.CosineSimilarity(recommendationsProfile[i], neighborsRecommendationsProfile[j])
				if err != nil {
					fmt.Println(err)
					continue
				}

				if newSim > sim {
					sim = newSim
				}
			}

			meanSim = +sim
		}

		sellability = sellability + (meanSim / float64(len(recommendationsProfile)))
	}

	//mean cosine similarity of all users
	sellability = sellability / float64(users.Nrow())
	return sellability
}
