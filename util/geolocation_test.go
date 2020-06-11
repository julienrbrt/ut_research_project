package util

import (
	"testing"
)

func TestDistanceTo(t *testing.T) {
	input := []Geolocation{
		{
			Latitude:  6.883238,
			Longitude: 52.210976,
		},
		{
			Latitude:  6.881124,
			Longitude: 52.216899,
		},
	}

	distExpected := 1.0
	distCalcultated := DistanceTo(input[0].Latitude, input[0].Longitude, input[1].Latitude, input[1].Longitude)

	if distCalcultated >= distExpected {
		t.Errorf("Distance is incorrect, got '%f', want '%f'", distCalcultated, distExpected)
	}

}

func TestBoundingCoordinates(t *testing.T) {
	input := []Geolocation{
		{
			Latitude:  6.883238,
			Longitude: 52.210976,
		},
		{
			Latitude:  6.881124,
			Longitude: 52.216899,
		},
	}

	distExpected := 1.0
	distCalcultated := BoundingCoordinates(input[0].Latitude, input[0].Longitude, distExpected)

	if distCalcultated[0].Latitude <= input[1].Latitude &&
		distCalcultated[0].Longitude <= input[1].Latitude &&
		distCalcultated[1].Latitude >= input[1].Latitude &&
		distCalcultated[1].Longitude >= input[1].Longitude {
		t.Errorf("Boundig Coordinates are incorrect")
	}
}
