package recommend

import (
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
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

//recipeCosineSimilarity calculates the cosine similarity of each recipes with each other
//Retuns the recipes id and their most similar recipes (sorted)
//CosineSimilarityRecipes uses tags and recipes ingredient
func recipeCosineSimilarity(recipes dataframe.DataFrame) (map[int][]int, error) {
	//keep only relevant columns
	var columnsToDrop []string
	for _, n := range recipes.Names() {
		//keep only orders columns and recipes tags
		if n != "id" && !strings.Contains(n, "tag_") && !strings.Contains(n, "ingredient_") {
			columnsToDrop = append(columnsToDrop, n)
		}
	}
	recipes = recipes.Drop(columnsToDrop)

	similarity := make(map[int][]int, recipes.Nrow())
	//unefficient but working
	for i, recipe := range recipes.Records() {
		//skip dataframe headers
		if i == 0 {
			continue
		}

		r, err := util.SS2SF(recipe[1:])
		if err != nil {
			return nil, err
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
				return nil, err
			}

			//calculate similarity of the two recipes
			sim, err := util.CosineSimilarity(r, ct)
			if err != nil {
				return nil, err
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

		var orderedRecipes []int
		for _, kv := range ss {
			k, err := strconv.Atoi(kv.Key)
			if err != nil {
				return nil, err
			}

			orderedRecipes = append(orderedRecipes, k)
		}

		recipeID, err := strconv.Atoi(recipe[0])
		if err != nil {
			return nil, err
		}

		log.Printf("%d / %d recipes cosine similarity calculated\n", i, recipes.Nrow())
		similarity[recipeID] = orderedRecipes
	}

	return similarity, nil
}

//WithContentFiltering recommends recipes using content filtering
func WithContentFiltering(userID, nbRecipes int, users, orders, recipes dataframe.DataFrame) error {
	log.Printf("(Content Filtering) Recommending Recipes for user %d", userID)

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
	recipeSim, err := recipeCosineSimilarity(recipes)
	if err != nil {
		return err
	}

	//select recipes to recommend
	recipesIDs, err := orders.Col("recipe_id").Int()
	if err != nil {
		return err
	}

	var recommendItems []int
	for _, r := range recipesIDs {
		recommendItems = append(recommendItems, recipeSim[r][:3]...)
	}

	//set maximum recommended recipes
	if len(recommendItems) > nbRecipes {
		recommendItems = recommendItems[:nbRecipes]
	}

	//get recipes names
	filters = []dataframe.F{}
	for _, i := range recommendItems {
		filters = append(filters, dataframe.F{Colname: "id", Comparator: series.Eq, Comparando: i})
	}
	recommendItemsName := recipes.FilterAggregation(dataframe.Or, filters...).Col("title").Records()
	for i, n := range recommendItemsName {
		log.Printf("%d/%d Recommend for user(%d) = %v\n", i+1, nbRecipes, userID, n)
	}

	return nil
}
