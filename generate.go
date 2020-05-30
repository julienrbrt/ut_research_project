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

// random food preferences for everythings that contain tag_

//GenerateUserData generates user data
func GenerateUserData(recipes dataframe.DataFrame) error {

	return nil
}
