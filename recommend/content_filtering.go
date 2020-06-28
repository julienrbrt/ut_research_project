package recommend

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
	"github.com/olekukonko/tablewriter"
	"gonum.org/v1/gonum/mat"
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

//interface gota with gonum
type matrix struct {
	dataframe.DataFrame
}

func (m matrix) At(i, j int) float64 {
	return m.Elem(i, j).Float()
}

func (m matrix) T() mat.Matrix {
	return mat.Transpose{Matrix: m}
}

//recipeSimilarityMatrix calculates the cosine similarity of each recipes with each other using recipes tags and ingredients
func recipeSimilarityMatrix(recipes dataframe.DataFrame) (*mat.Dense, error) {
	//keep only relevant columns
	var columnsToDrop []string
	for _, n := range recipes.Names() {
		//keep only orders columns and recipes tags
		if n != "id" && !strings.Contains(n, "tag_") && !strings.Contains(n, "ingredient_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}
	recipes = recipes.Drop(columnsToDrop)

	//create similarity matrix
	matrix := mat.NewDense(recipes.Nrow(), recipes.Nrow(), nil)

	for i, recipe := range recipes.Records()[1:] {
		r, err := util.SS2SF(recipe[1:])
		if err != nil {
			return &mat.Dense{}, err
		}

		for j, compareTo := range recipes.Records()[1:] {
			ct, err := util.SS2SF(compareTo[1:])
			if err != nil {
				return &mat.Dense{}, err
			}

			//calculate similarity of the two recipes
			sim, err := util.CosineSimilarity(r, ct)
			if err != nil {
				return &mat.Dense{}, err
			}

			//set value
			matrix.Set(i, j, sim)
		}

		log.Printf("%d / %d recipes cosine similarity calculated\n", i+1, recipes.Nrow())
	}

	// df := dataframe.LoadMatrix(matrix)
	// util.WriteCSV(df, "data/recipes_matrix.csv")

	return matrix, nil
}

func recommendedContentFiltering(userID, nbRecipes, nbTags int, km float64, users, orders, recipes dataframe.DataFrame) ([]float64, error) {
	//user profile
	orders = userProfileOrder(userID, orders, recipes)
	log.Printf("User %d has made %d orders with a (normalized) average rating of %.2f per order\n", userID, orders.Nrow(), orders.Col("rating").Mean())

	//calculate user preferences tags weight
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

	//recommend from the nbTags most liked tags
	if len(tags) >= nbTags {
		tags = tags[:nbTags]
	}

	//filter tags
	var filters []dataframe.F
	for i := range tags {
		filters = append(filters, dataframe.F{
			Colname: tags[i], Comparator: series.Eq, Comparando: "1",
		})
	}

	//get orders that match the tags and sort by prefered orders (highest rating)
	orders = orders.FilterAggregation(dataframe.Or, filters...).Arrange(dataframe.RevSort("rating"))
	//select recipes to recommend
	ids, err := orders.Col("recipe_id").Int()
	if err != nil {
		return nil, err
	}
	if len(ids) >= nbTags {
		ids = ids[:nbTags]
	}

	//calculate cosine similarity
	// sim, err := recipeSimilarityMatrix(recipes)
	// if err != nil {
	// 	return nil, err
	// }
	sim := matrix{util.LoadCSV("data/recipes_matrix.csv")}

	var recommendItems []float64
	for _, r := range ids {
		//r-1 because id starts at 1 but matrix at 0
		bestMatch := mat.Row(nil, r-1, sim)
		sort.Slice(bestMatch, func(i, j int) bool {
			return bestMatch[i] > bestMatch[j]
		})

		recommendItems = append(recommendItems, bestMatch[:nbTags]...)
	}

	//set maximum recommended recipes
	if len(recommendItems) > nbRecipes {
		recommendItems = recommendItems[:nbRecipes]
	}

	return recommendItems, nil
}

//WithContentFiltering recommends recipes using content filtering
//returns the recommended recipes_id
func WithContentFiltering(userID, nbRecipes, nbTags int, km float64, users, orders, recipes dataframe.DataFrame) error {
	log.Printf("(Content Filtering) Recommending Recipes for user %d", userID)

	//keep only neighboring users
	neighborsUsers := usersCloseByXKm(userID, km, users)

	//calculate recommended recipes
	recommendItems, err := recommendedContentFiltering(userID, nbRecipes, nbTags, km, users, orders, recipes)
	if err != nil {
		return err
	}

	//calculate sellability
	sellability := MeasureContentSellability(userID, nbRecipes, nbTags, km, recommendItems, neighborsUsers, orders, recipes)

	//fill in table with scores and recommended items
	lines := make([][]string, 0)
	lines = append(lines, []string{
		fmt.Sprint("Content Filtering"),  //model
		fmt.Sprintf("%.5f", 0.0),         //precision@nbRecipes
		fmt.Sprintf("%.5f", 0.0),         //recall@NbRecipes
		fmt.Sprintf("%.5f", sellability), //sellability@km
		fmt.Sprintf("%v", recommendItems),
	})

	//print table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Model", fmt.Sprintf("Precision@%d", nbRecipes), fmt.Sprintf("Recall@%d", nbRecipes),
		fmt.Sprintf("Sellability@%.5f", km), "Recommendation"})
	for _, v := range lines {
		table.Append(v)
	}
	table.Render()

	return nil
}
