package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"

	"github.com/brianvoe/gofakeit/v5"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

const usersCSVPath = "data/users_data.csv"
const ordersCSVPath = "data/orders_data.csv"

//User contains the food preferences data of an user
type User struct {
	ID              int
	Name            string
	Age             int
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

//generateUsers generates user data
func generateUsers(n int, df dataframe.DataFrame) (GeneratedUsers, error) {
	var result GeneratedUsers
	var err error

	//build tags list
	reg, err := regexp.Compile("tag_")
	if err != nil {
		return GeneratedUsers{}, err
	}

	tags := make(map[int]string)
	for i, t := range df.Names() {
		if reg.MatchString(t) {
			tags[i] = t
		}
	}

	for i := 0; i < n; i++ {
		var user User

		user.ID = i
		user.Name = gofakeit.Name()
		user.Age = gofakeit.Number(18, 70)
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
		// assumption that a meal-sharing user will not enter more than 8 tags
		randomNb := rand.Intn(8)
		//get subset of tags preferences
		for i := 0; i < randomNb; i++ {
			user.FoodPreferences = append(user.FoodPreferences, tags[randomTags[i]])
		}

		//random number of orders that match food preferences
		//assumption that a meal-sharing user will not have more than 35 orders
		randomNb = rand.Intn(35)

		//match recipe of user food prefrences
		var filterRecipes []dataframe.F
		for _, f := range user.FoodPreferences {
			filterRecipes = append(filterRecipes, dataframe.F{Colname: f, Comparator: series.Eq, Comparando: "1"})
		}
		matchingRecipes := df.Copy().Filter(filterRecipes...).Records()

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

			//add recipes that user bought and didn't match its taste
			if rand.Intn(100) >= 85 {
				orderID = rand.Intn(df.Nrow())
				//randomness give us a recipes out of bound, skip this iteration
				if orderID == 0 || orderID >= len(matchingRecipes) {
					continue
				}
			}

			user.OrdersHistory = append(user.OrdersHistory, orderID)
		}

		//random rating from order history, matching mean rating of the recipe
		for range user.OrdersHistory {
			//we generate rating that includes 0 as we consider those grades as recipe unrated
			user.OrdersRating = append(user.OrdersRating, rand.Intn(5))
		}

		result.Users = append(result.Users, user)
	}

	return result, nil
}

//asUserDf converts a list of user as a dataframe
func (users *GeneratedUsers) asUserDf(tags []string) dataframe.DataFrame {
	log.Println("Processing...")

	headers := []string{"id", "name", "age", "latitude", "longitude"}
	records := [][]string{}

	//append tags to headers
	// headers = append(headers, tags...)

	for _, user := range users.Users {
		//fill in data
		data := []string{
			strconv.Itoa(user.ID),
			user.Name,
			strconv.Itoa(user.Age),
			fmt.Sprintf("%f", user.Latitude),
			fmt.Sprintf("%f", user.Longitude),
		}

		records = append(records, data)
	}

	//load as dataframe
	df := dataframe.LoadRecords(append([][]string{headers}, records...))

	return df
}

//asOrderDf converts build the orders history and rating dataframe
func (users *GeneratedUsers) asOrderDf() dataframe.DataFrame {
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

//LoadUsersData generate N user data
func LoadUsersData(n int, recipes dataframe.DataFrame, writeCSV bool) (dataframe.DataFrame, dataframe.DataFrame) {
	//generate data
	users, err := generateUsers(n, recipes)
	if err != nil {
		log.Fatalln(err)
	}

	//load food preferences (aka tags)
	reg, err := regexp.Compile("tag_")
	if err != nil {
		log.Fatalln(err)
	}

	var tags []string
	for _, t := range recipes.Names() {
		if reg.MatchString(t) {
			tags = append(tags, t)
		}
	}

	//processing and load data
	usersDF := users.asUserDf(tags)
	ordersDF := users.asOrderDf()

	if writeCSV {
		//save user csv
		if err := WriteCSV(usersDF, usersCSVPath); err != nil {
			log.Fatalln(err)
		}

		//save order csv
		if err := WriteCSV(ordersDF, ordersCSVPath); err != nil {
			log.Fatalln(err)
		}
	}

	return usersDF, ordersDF
}
