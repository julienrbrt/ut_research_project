package recommend

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strconv"

	"github.com/go-gota/gota/dataframe"
	"github.com/olekukonko/tablewriter"
	"github.com/zhenghaoz/gorse/base"
	"github.com/zhenghaoz/gorse/core"
	"github.com/zhenghaoz/gorse/model"
)

var models = []core.ModelInterface{
	//BaseLine
	model.NewBaseLine(base.Params{
		base.NEpochs: 10,
	}),
	// SlopOne
	model.NewSlopOne(nil),
	// CoClustering
	model.NewCoClustering(base.Params{
		base.NUserClusters: 5,
		base.NItemClusters: 5,
		base.NEpochs:       10,
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
		base.NEpochs:    10,
		base.Reg:        0.1,
		base.Lr:         0.01,
		base.NFactors:   50,
		base.InitMean:   0,
		base.InitStdDev: 0.001,
	}),
	//BRP
	model.NewBPR(base.Params{
		base.NFactors:   50,
		base.Reg:        0.005,
		base.Lr:         0.01,
		base.NEpochs:    50,
		base.InitMean:   0,
		base.InitStdDev: 0.001,
	}),
}

//WithCollaborativeFiltering recommends recipes using collaborative filtering
func WithCollaborativeFiltering(userID, nbRecipes int, km float64, users, orders, recipes dataframe.DataFrame) error {
	log.Printf("(Collaborative Filtering) Recommending Recipes for user %d", userID)

	//keep only neighboring users
	neighborsUsers := usersCloseByXKm(userID, km, users)

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
		scoresRating := core.EvaluateRating(m, test, core.RMSE)
		//generate recommendations for user
		//get all items in the full dataset
		items := core.Items(data)
		//get user ratings in the training set
		excludeItems := train.User(strconv.Itoa(userID))
		//get top recommended items (excluding rated items)
		recommendItems, _ := core.Top(items, strconv.Itoa(userID), nbRecipes, excludeItems, m)

		//calculate sellability
		sellability := MeasureCollaborativeSellability(userID, nbRecipes, km, recommendItems, m, data, train, test, neighborsUsers)

		//fill in table with scores and recommended items
		lines = append(lines, []string{
			fmt.Sprint(reflect.TypeOf(m)),         //model
			fmt.Sprintf("%.5f", scoresRanking[0]), //precision@nbRecipes
			fmt.Sprintf("%.5f", scoresRanking[1]), //recall@NbRecipes
			fmt.Sprintf("%.5f", scoresRating[0]),  //rmse@nbRecipes
			fmt.Sprintf("%.5f", sellability),      //sellability@neighborsUsers.Nrow()
			fmt.Sprintf("%v", recommendItems),
		})
	}

	//print table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Model", fmt.Sprintf("Precision@%d", nbRecipes), fmt.Sprintf("Recall@%d", nbRecipes),
		fmt.Sprintf("RMSE@%d", nbRecipes), fmt.Sprintf("Sellability@%d", neighborsUsers.Nrow()), "Recommendation"})
	for _, v := range lines {
		table.Append(v)
	}
	table.Render()

	return nil
}

//BestHyperParametersKNN test best fitting parameters of the KNN model
func BestHyperParametersKNN(nbRecipes int, orders dataframe.DataFrame) {
	//load dataset
	data := core.NewDataSet(orders.Col("user_id").Records(), orders.Col("recipe_id").Records(), orders.Col("rating").Float())

	cv := core.GridSearchCV(model.NewKNN(nil), data, core.ParameterGrid{
		base.Lr:         {0.005, 0.05, 0.1},
		base.Reg:        {0.005, 0.02, 0.5},
		base.NEpochs:    {50},
		base.Type:       {base.Baseline},
		base.Similarity: {base.Cosine, base.MSD},
		base.K:          {10, 40, 80},
	}, core.NewKFoldSplitter(5), 0, &base.RuntimeOptions{false, 1, runtime.NumCPU()},
		core.NewRankEvaluator(nbRecipes, core.Precision, core.Recall), core.NewRatingEvaluator(core.RMSE))
	fmt.Println("=== Precision@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[0].BestScore)
	fmt.Printf("The best params is: %v\n", cv[0].BestParams)
	fmt.Println("=== Recall@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[1].BestScore)
	fmt.Printf("The best params is: %v\n", cv[1].BestParams)
	fmt.Println("=== RMSE@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[2].BestScore)
	fmt.Printf("The best params is: %v\n", cv[2].BestParams)
}

//BestHyperParametersBaseLine test best fitting parameters of the BaseLine model
func BestHyperParametersBaseLine(nbRecipes int, orders dataframe.DataFrame) {
	//load dataset
	data := core.NewDataSet(orders.Col("user_id").Records(), orders.Col("recipe_id").Records(), orders.Col("rating").Float())

	cv := core.GridSearchCV(model.NewKNN(nil), data, core.ParameterGrid{
		base.Lr:      {0.005, 0.05, 0.1},
		base.Reg:     {0.005, 0.02, 0.5},
		base.NEpochs: {50},
	}, core.NewKFoldSplitter(5), 0, &base.RuntimeOptions{false, 1, runtime.NumCPU()},
		core.NewRankEvaluator(nbRecipes, core.Precision, core.Recall), core.NewRatingEvaluator(core.RMSE))
	fmt.Println("=== Precision@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[0].BestScore)
	fmt.Printf("The best params is: %v\n", cv[0].BestParams)
	fmt.Println("=== Recall@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[1].BestScore)
	fmt.Printf("The best params is: %v\n", cv[1].BestParams)
	fmt.Println("=== RMSE@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[2].BestScore)
	fmt.Printf("The best params is: %v\n", cv[2].BestParams)
}

//BestHyperParametersBPR test best fitting parameters of the BPR model
func BestHyperParametersBPR(nbRecipes int, orders dataframe.DataFrame) {
	//load dataset
	data := core.NewDataSet(orders.Col("user_id").Records(), orders.Col("recipe_id").Records(), orders.Col("rating").Float())

	cv := core.GridSearchCV(model.NewBPR(nil), data, core.ParameterGrid{
		base.NFactors:   {5, 10, 50},
		base.Reg:        {0.005, 0.01, 0.5},
		base.Lr:         {0.01, 0.05, 0.1},
		base.NEpochs:    {50},
		base.InitMean:   {0},
		base.InitStdDev: {0.001},
	}, core.NewKFoldSplitter(5), 0, &base.RuntimeOptions{false, 1, runtime.NumCPU()},
		core.NewRankEvaluator(nbRecipes, core.Precision, core.Recall), core.NewRatingEvaluator(core.RMSE))
	fmt.Println("=== Precision@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[0].BestScore)
	fmt.Printf("The best params is: %v\n", cv[0].BestParams)
	fmt.Println("=== Recall@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[1].BestScore)
	fmt.Printf("The best params is: %v\n", cv[1].BestParams)
	fmt.Println("=== RMSE@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[2].BestScore)
	fmt.Printf("The best params is: %v\n", cv[2].BestParams)
}

//BestHyperParametersSVD test best fitting parameters of the SVD model
func BestHyperParametersSVD(nbRecipes int, orders dataframe.DataFrame) {
	//load dataset
	data := core.NewDataSet(orders.Col("user_id").Records(), orders.Col("recipe_id").Records(), orders.Col("rating").Float())

	cv := core.GridSearchCV(model.NewSVD(nil), data, core.ParameterGrid{
		base.NFactors:   {5, 10, 50, 100},
		base.Reg:        {0.005, 0.2, 0.5},
		base.Lr:         {0.005, 0.05, 0.1},
		base.NEpochs:    {50},
		base.InitMean:   {0},
		base.InitStdDev: {0.001, 0.1},
	}, core.NewKFoldSplitter(5), 0, &base.RuntimeOptions{false, 1, runtime.NumCPU()},
		core.NewRankEvaluator(nbRecipes, core.Precision, core.Recall), core.NewRatingEvaluator(core.RMSE))
	fmt.Println("=== Precision@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[0].BestScore)
	fmt.Printf("The best params is: %v\n", cv[0].BestParams)
	fmt.Println("=== Recall@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[1].BestScore)
	fmt.Printf("The best params is: %v\n", cv[1].BestParams)
	fmt.Println("=== RMSE@nbRecipes")
	fmt.Printf("The best score is: %.5f\n", cv[2].BestScore)
	fmt.Printf("The best params is: %v\n", cv[2].BestParams)
}
