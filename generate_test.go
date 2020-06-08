package main

import (
	"testing"

	"github.com/brianvoe/gofakeit/v5"
)

func TestGenerateUsers(t *testing.T) {
	//set seed
	gofakeit.Seed(42)

	recipes := LoadCSV(recipesCSVPath)
	users, err := generateUsers(15, recipes)
	if err != nil {
		panic(err)
	}

	if len(users.Users) != 15 {
		t.Errorf("Number of users is incorrect, got '%d', want '%d'", len(users.Users), 15)
	}

	for _, u := range users.Users {
		//verify location
		if u.Latitude > maxLatitudeNL {
			t.Errorf("Latidude is incorrect, got '%v', shoud be maximum '%v'", u.Latitude, maxLatitudeNL)
		}
		if u.Latitude < minLatitudeNL {
			t.Errorf("Latidude is incorrect, got '%v', shoud be minimum '%v'", u.Latitude, minLatitudeNL)

		}
		if u.Longitude > maxLongitudeNL {
			t.Errorf("Latidude is incorrect, got '%v', shoud be maximum '%v'", u.Longitude, maxLongitudeNL)
		}
		if u.Longitude < minLongitudeNL {
			t.Errorf("Latidude is incorrect, got '%v', shoud be minimum '%v'", u.Longitude, minLongitudeNL)
		}
	}
}
