package recommend

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
	"github.com/olekukonko/tablewriter"
)

//userProfileOrder returns a dataframe containing orders and their normalized rating
//appends as well the tags description columns to the orders dataframe
func userProfileOrder(userID int, orders, recipes dataframe.DataFrame) dataframe.DataFrame {
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

//userTagsWeight will use the UserProfile output to generate the recipes tags weight based on the ratings of each recipes using that tag
func userTagsWeight(orders dataframe.DataFrame) map[string]float64 {
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

type kv struct {
	Key   string
	Value float64
}

type sim struct {
	RecipeID  int
	SimilarTo string
}

//recipeCosineSimilarity calculates the cosine similarity of each recipes with each other
//Retuns the recipes id and their most similar recipes (sorted by most relevant first)
//CosineSimilarityRecipes uses tags and recipes ingredient
func recipeCosineSimilarity(recipes dataframe.DataFrame) (dataframe.DataFrame, error) {
	//keep only relevant columns
	var columnsToDrop []string
	for _, n := range recipes.Names() {
		//keep only orders columns and recipes tags
		if n != "id" && !strings.Contains(n, "tag_") && !strings.Contains(n, "ingredient_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}
	recipes = recipes.Drop(columnsToDrop)

	similarities := []sim{}
	//unefficient but working
	for i, recipe := range recipes.Records() {
		//skip dataframe headers
		if i == 0 {
			continue
		}

		r, err := util.SS2SF(recipe[1:])
		if err != nil {
			return dataframe.DataFrame{}, err
		}

		//holds the compared recipes id and it's similarity
		similarTo := make(map[string]float64)
		for j, compareTo := range recipes.Records() {
			//skip dataframe headers and identical rows
			if j == 0 || i == j {
				continue
			}

			ct, err := util.SS2SF(compareTo[1:])
			if err != nil {
				return dataframe.DataFrame{}, err
			}

			//calculate similarity of the two recipes
			sim, err := util.CosineSimilarity(r, ct)
			if err != nil {
				return dataframe.DataFrame{}, err
			}

			similarTo[compareTo[0]] = sim
		}

		//sort the most similar recipes
		var ss []kv
		for k, v := range similarTo {
			ss = append(ss, kv{k, v})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Value > ss[j].Value
		})

		//should be better to be saved in []int but this would not support a dataframe
		var orderedRecipes string
		for _, kv := range ss {
			orderedRecipes = orderedRecipes + " " + kv.Key
		}
		orderedRecipes = strings.Trim(orderedRecipes, " ")

		recipeID, err := strconv.Atoi(recipe[0])
		if err != nil {
			return dataframe.DataFrame{}, err
		}

		log.Printf("%d / %d recipes cosine similarity calculated\n", i, recipes.Nrow())
		similarities = append(similarities, sim{RecipeID: recipeID, SimilarTo: orderedRecipes})
	}

	return dataframe.LoadStructs(similarities), nil
}

func recommendedContentFiltering(userID, nbRecipes int, km float64, users, orders, recipes dataframe.DataFrame) ([]string, error) {
	//user profile
	orders = userProfileOrder(userID, orders, recipes)
	log.Printf("User %d has made %d orders with a (normalized) average rating of %.2f per order\n", userID, orders.Nrow(), orders.Col("rating").Mean())

	//calculate users preference tags weight
	tagsWeight := userTagsWeight(orders)
	//sort the most prefered tags
	var tw []kv
	for k, v := range tagsWeight {
		tw = append(tw, kv{k, v})
	}
	sort.Slice(tw, func(i, j int) bool {
		return tw[i].Value > tw[j].Value
	})

	var tags []string
	for _, kv := range tw {
		tags = append(tags, kv.Key)
	}

	//recommend from the 3 most liked tags
	if len(tags) >= 3 {
		tags = tags[:3]
	}

	//filter tags
	var filters []dataframe.F
	for i := range tags {
		filters = append(filters, dataframe.F{
			Colname: tags[i], Comparator: series.Eq, Comparando: "1",
		})
	}

	//get placed orders that match that tags and sort by prefered orders (highest rating)
	orders = orders.FilterAggregation(dataframe.Or, filters...).Arrange(dataframe.RevSort("rating"))

	//calculate cosine similarity
	log.Println("Calculating recipes cosine similarity...")
	// recipesSim, err := recipeCosineSimilarity(recipes)
	// if err != nil {
	// 	return err
	// }
	recipesSim := util.LoadCSV("data/recipes_sim.csv")

	//select recipes to recommend
	recipesIDs, err := orders.Col("recipe_id").Int()
	if err != nil {
		return nil, err
	}

	var recommendItems []string
	for _, r := range recipesIDs {
		match := recipesSim.
			Filter(dataframe.F{Colname: "RecipeID", Comparator: series.Eq, Comparando: r}).
			Col("SimilarTo").
			String()
		recommendItems = append(recommendItems, strings.Split(strings.Trim(match, "[ ]"), " ")[:3]...)
	}

	//set maximum recommended recipes
	if len(recommendItems) > nbRecipes {
		recommendItems = recommendItems[:nbRecipes]
	}

	return recommendItems, nil
}

//WithContentFiltering recommends recipes using content filtering
//returns the recommended recipes_id
func WithContentFiltering(userID, nbRecipes int, km float64, users, orders, recipes dataframe.DataFrame) error {
	log.Printf("(Content Filtering) Recommending Recipes for user %d", userID)

	//keep only neighboring users
	neighborsUsers := usersCloseByXKm(userID, km, users)

	//calculate recommended recipes
	recommendItems, err := recommendedContentFiltering(userID, nbRecipes, km, users, orders, recipes)
	if err != nil {
		return err
	}

	//calculate sellability
	sellability := MeasureContentSellability(userID, nbRecipes, km, recommendItems, neighborsUsers, orders, recipes)

	//fill in table with scores and recommended items
	lines := make([][]string, 0)
	lines = append(lines, []string{
		fmt.Sprint("Content Filtering"),  //model
		fmt.Sprintf("%.5f", 0.0),         //precision@nbRecipes
		fmt.Sprintf("%.5f", 0.0),         //recall@NbRecipes
		fmt.Sprintf("%.5f", 0.0),         //rmse@nbRecipes
		fmt.Sprintf("%.5f", sellability), //sellability@neighborsUsers.Nrow()
		fmt.Sprintf("%v", recommendItems),
	})

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
