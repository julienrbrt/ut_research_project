package main

import "github.com/go-gota/gota/dataframe"

//User contains the food preferences data of an user
type User struct {
	Name            string
	Age             string
	Latitude        float32
	Longitude       float32
	FoodPreferences []string
	OrdersHistory   []string
	OrdersRating    map[string]int
}

//FoodPreferences contains the user food preferences
type FoodPreferences string

const (
	vegetarian FoodPreferences = "vegetarisch"
	vegan                      = "vegan"
	budget                     = "budget"
	snack                      = "frituren"
	french                     = "frans"
	spanish                    = "spaans"
)

//GenerateUserData generates user data
func GenerateUserData(recipes dataframe.DataFrame) error {

	return nil
}
