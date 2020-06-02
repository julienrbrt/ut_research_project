package main

import (
	"fmt"

	"github.com/go-gota/gota/dataframe"
	"github.com/zhenghaoz/gorse/base"
	"github.com/zhenghaoz/gorse/core"
	"github.com/zhenghaoz/gorse/model"
)

// orders = feedback, users = users, recipes = items

//BPRRecommender is a recommender using the BPR algorithm
func BPRRecommender(orders, users, recipes dataframe.DataFrame) {
	//load dataset
	data := core.LoadDataFromCSV(ordersCSVPath, ",", true)
	//split dataset
	train, test := core.Split(data, 0.2)
	//create model
	bpr := model.NewBPR(base.Params{
		base.NFactors:   10,
		base.Reg:        0.01,
		base.Lr:         0.05,
		base.NEpochs:    1000,
		base.InitMean:   0,
		base.InitStdDev: 0.001,
	})
	// Fit model
	bpr.Fit(train, nil)
	// Evaluate model
	scores := core.EvaluateRank(bpr, test, train, 10, core.Precision, core.Recall, core.NDCG)
	fmt.Printf("Precision@10 = %.5f\n", scores[0])
	fmt.Printf("Recall@10 = %.5f\n", scores[1])
	fmt.Printf("NDCG@10 = %.5f\n", scores[1])
	// Generate recommendations for user(4):
	// Get all items in the full dataset
	items := core.Items(data)
	// Get user(55)'s ratings in the training dataset
	excludeItems := train.User("55")
	// Get top 3 recommended items (excluding rated items) for user(55) using BPR
	recommendItems, _ := core.Top(items, "55", 3, excludeItems, bpr)
	fmt.Printf("Recommend for user(55) = %v\n", recommendItems)
}
