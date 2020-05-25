package main

//User contains the food preferences data of an user
type User struct {
	Name            string
	Age             string
	Latitude        float32
	Longitude       float32
	OrdersRating    map[string]int
	FoodPreferences []string
}

//GenerateUserData generates user data
func GenerateUserData(recipes *AHRecipes) error {

	return nil
}
