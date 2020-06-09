package generate

import (
	"fmt"
	"log"
	"math/rand"
	"path"
	"regexp"
	"strconv"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/julienrbrt/ut_research_project/util"
	"github.com/kardianos/osext"
)

//UsersCSVPath CSV path of the dataset users
const UsersCSVPath = "data/users.csv"

//OrdersCSVPath CSV path of the dataset orders
const OrdersCSVPath = "data/orders.csv"

//User contains the food preferences data of an user
type User struct {
	ID              int
	Name            string
	Latitude        float64
	Longitude       float64
	FoodPreferences []string
	OrdersHistory   []int
	OrdersRating    []int
}

//GeneratedUsers contains a list of generated users
type GeneratedUsers struct {
	Users []User
}

//Netherlands bounding box https://boundingbox.klokantech.com/
const minLatitudeNL = 51.1028
const maxLatitudeNL = 53.5842
const minLongitudeNL = 3.987
const maxLongitudeNL = 7.8929

//generateUsers generates user data with recipes
func generateUsers(n int, recipes dataframe.DataFrame) (GeneratedUsers, error) {
	var users GeneratedUsers
	var err error

	//build tags list
	reg, err := regexp.Compile("tag_")
	if err != nil {
		return GeneratedUsers{}, err
	}

	tags := make(map[int]string)
	for i, t := range recipes.Names() {
		if reg.MatchString(t) {
			tags[i] = t
		}
	}

	for i := 0; i < n; i++ {
		log.Printf("Generating user %d / %d...\n", i+1, n)

		var user User

		user.ID = i + 1
		user.Name = gofakeit.Name()
		user.Latitude, err = gofakeit.LatitudeInRange(minLatitudeNL, maxLatitudeNL)
		if err != nil {
			return GeneratedUsers{}, err
		}
		user.Longitude, err = gofakeit.LongitudeInRange(minLongitudeNL, maxLongitudeNL)
		if err != nil {
			return GeneratedUsers{}, err
		}

		//random food preferences (aka tags)
		randomTags := rand.Perm(len(tags))
		// assumption that a meal-sharing user will not enter more than 12 tags
		randomNb := rand.Intn(12)
		//get subset of tags preferences
		for i := 0; i < randomNb; i++ {
			user.FoodPreferences = append(user.FoodPreferences, tags[randomTags[i]])
		}

		//random number of orders that match food preferences
		//assumption that a meal-sharing user will not have more than 55 orders
		randomNb = rand.Intn(50) + 5

		//match recipe of user food prefrences
		var filterRecipes []dataframe.F
		for _, f := range user.FoodPreferences {
			filterRecipes = append(filterRecipes, dataframe.F{Colname: f, Comparator: series.Eq, Comparando: "1"})
		}
		matchingRecipes := recipes.Copy().Filter(filterRecipes...).Records()

		//add recipes matching user taste
		for i := 0; i < randomNb; i++ {
			//random index
			index := rand.Intn(randomNb)
			//randomness give us a recipes out of bound, skip this iteration
			if index == 0 || index >= len(matchingRecipes) {
				continue
			}

			//generate random orderID
			orderID, err := strconv.Atoi(matchingRecipes[index][0])
			if err != nil {
				continue
			}

			user.OrdersHistory = append(user.OrdersHistory, orderID)
		}

		//random rating from order history, matching mean rating of the recipe
		for range user.OrdersHistory {
			//we generate order rating
			user.OrdersRating = append(user.OrdersRating, rand.Intn(5)+1)
		}

		users.Users = append(users.Users, user)
	}

	return users, nil
}

//transformToUserDF converts a list of user as a dataframe
func (users *GeneratedUsers) transformToUserDF(tags []string) dataframe.DataFrame {
	log.Println("Processing...")

	headers := []string{"id", "name", "latitude", "longitude"}
	records := [][]string{}

	//append tags to headers
	headers = append(headers, tags...)

	for _, user := range users.Users {
		//fill in data
		data := []string{
			strconv.Itoa(user.ID),
			user.Name,
			fmt.Sprintf("%f", user.Latitude),
			fmt.Sprintf("%f", user.Longitude),
		}

		//map of contained tags
		set := make(map[string]bool)
		for _, t := range user.FoodPreferences {
			set[t] = true
		}

		//add tags
		for _, t := range tags {
			if set[t] {
				data = append(data, "1")
			} else {
				data = append(data, "0")
			}
		}

		records = append(records, data)
	}

	//load as dataframe
	df := dataframe.LoadRecords(append([][]string{headers}, records...))

	return df
}

//transformToOrderDF converts build the orders history and rating dataframe
func (users *GeneratedUsers) transformToOrderDF() dataframe.DataFrame {
	log.Println("Processing...")

	headers := []string{"user_id", "order_id", "rating"}
	records := [][]string{}

	for _, user := range users.Users {
		for i := range user.OrdersHistory {
			//fill in data
			data := []string{
				strconv.Itoa(user.ID),
				strconv.Itoa(user.OrdersHistory[i]),
				strconv.Itoa(user.OrdersRating[i]),
			}

			records = append(records, data)
		}
	}

	//load as dataframe
	df := dataframe.LoadRecords(append([][]string{headers}, records...))

	return df
}

//UsersData generate N user data
func UsersData(n int, recipes dataframe.DataFrame, writeCSV bool) error {
	//generate data
	users, err := generateUsers(n, recipes)
	if err != nil {
		return err
	}

	//load food preferences (aka tags)
	reg, err := regexp.Compile("tag_")
	if err != nil {
		return err
	}

	var tags []string
	for _, t := range recipes.Names() {
		if reg.MatchString(t) {
			tags = append(tags, t)
		}
	}

	//processing and load data
	usersDF := users.transformToUserDF(tags)
	ordersDF := users.transformToOrderDF()

	if writeCSV {
		//get executable flolder
		ef, err := osext.ExecutableFolder()
		if err != nil {
			return err
		}

		//save user csv
		if err := util.WriteCSV(usersDF, path.Join(ef, UsersCSVPath)); err != nil {
			return err
		}

		//save order csv
		if err := util.WriteCSV(ordersDF, path.Join(ef, OrdersCSVPath)); err != nil {
			return err
		}
	}

	return nil
}
