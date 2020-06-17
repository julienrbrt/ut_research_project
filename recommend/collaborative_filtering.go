package recommend

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
	"github.com/olekukonko/tablewriter"
	"github.com/zhenghaoz/gorse/base"
	"github.com/zhenghaoz/gorse/core"
	"github.com/zhenghaoz/gorse/model"
)

//usersCloseByXKm returns a dataframe containings users around user with userID from x kms
func usersCloseByXKm(usersID int, km float64, users dataframe.DataFrame) dataframe.DataFrame {
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

var models = []core.ModelInterface{
	//BaseLine
	model.NewBaseLine(base.Params{
		base.NEpochs: 1000,
	}),
	// SlopOne
	model.NewSlopOne(nil),
	// CoClustering
	model.NewCoClustering(base.Params{
		base.NUserClusters: 5,
		base.NItemClusters: 3,
		base.NEpochs:       1000,
	}),
	// KNN
	model.NewKNN(base.Params{
		base.Type:       base.Baseline,
		base.UserBased:  false,
		base.Similarity: base.Pearson,
		base.K:          30,
		base.Shrinkage:  90,
	}),
	// SVD
	model.NewSVD(base.Params{
		base.NEpochs:    1000,
		base.Reg:        0.1,
		base.Lr:         0.01,
		base.NFactors:   50,
		base.InitMean:   0,
		base.InitStdDev: 0.001,
	}),
	//SVD++
	model.NewSVDpp(base.Params{
		base.NEpochs:    1000,
		base.Reg:        0.05,
		base.Lr:         0.005,
		base.NFactors:   50,
		base.InitMean:   0,
		base.InitStdDev: 0.001,
	}),
	//ItemPop
	model.NewItemPop(nil),
	//KNN (Implicit)
	model.NewKNNImplicit(nil),
	//BRP
	model.NewBPR(base.Params{
		base.NFactors:   10,
		base.Reg:        0.01,
		base.Lr:         0.05,
		base.NEpochs:    1000,
		base.InitMean:   0,
		base.InitStdDev: 0.001,
	}),
	///WRMF
	model.NewWRMF(base.Params{
		base.NFactors: 20,
		base.Reg:      0.015,
		base.Alpha:    1.0,
		base.NEpochs:  1000,
	}),
}

//WithCollaborativeFiltering recommends recipes using collaborative filtering
func WithCollaborativeFiltering(userID, nbRecipes int, km float64, users, orders, recipes dataframe.DataFrame) (map[string][]string, error) {
	log.Printf("(Collaborative Filtering) Recommending Recipes for user %d", userID)

	//filter users table by neighbors users
	users = usersCloseByXKm(userID, km, users)
	//filter orders rating from neighbors users
	orders = orders.InnerJoin(users.Copy().Rename("user_id", "id"), "user_id")

	if orders.Nrow() == 0 {
		return nil, errors.New("No neighbors users, sellability is then not taken in account")
	}
	log.Printf("All (neighbors) users are in number of %d and have made %d orders\n", users.Nrow(), orders.Nrow())

	//load dataset
	data := core.NewDataSet(orders.Col("user_id").Records(), orders.Col("recipe_id").Records(), orders.Col("rating").Float())
	//split dataset
	train, test := core.Split(data, 0.2)

	//create model
	lines := make([][]string, 0)
	for _, m := range models {
		//fit model
		m.Fit(train, nil)
		//evaluate model
		scoresRanking := core.EvaluateRank(m, test, train, nbRecipes, core.Precision, core.Recall)
		scoresRating := core.EvaluateRating(m, test, core.RMSE, core.MAE)
		//generate recommendations for user
		//get all items in the full dataset
		items := core.Items(data)
		//get user ratings in the training set
		excludeItems := train.User(strconv.Itoa(userID))
		//get top recommended items (excluding rated items)
		recommendItems, _ := core.Top(items, strconv.Itoa(userID), nbRecipes, excludeItems, m)
		//fill in table with scores and recommended items
		lines = append(lines, []string{
			fmt.Sprint(reflect.TypeOf(m)),         //model
			fmt.Sprintf("%.5f", scoresRanking[0]), //precision@nbRecipes
			fmt.Sprintf("%.5f", scoresRanking[1]), //recall@NbRecipes
			fmt.Sprintf("%.5f", scoresRating[0]),  //rmse@nbRecipes
			fmt.Sprintf("%.5f", scoresRating[1]),  //mae@NbRecipes
			fmt.Sprintf("%v", recommendItems),
		})
	}

	//print table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Model", fmt.Sprintf("Precision@%d", nbRecipes), fmt.Sprintf("Recall@%d", nbRecipes),
		fmt.Sprintf("RMSE@%d", nbRecipes), fmt.Sprintf("MAE@%d", nbRecipes), "Recommendation"})
	for _, v := range lines {
		table.Append(v)
	}
	table.Render()

	recommendItems := make(map[string][]string, len(lines))
	for _, l := range lines {
		recommendItems[l[0]] = strings.Split(strings.Trim(l[5], "[ ]"), " ")
	}

	return recommendItems, nil
}
