package util

import "math"

//http://janmatuschek.de/LatitudeLongitudeBoundingCoordinates

//Geolocation contains coordinates of a postion
type Geolocation struct {
	Latitude  float64
	Longitude float64
}

const earthRadius = 6371.01

//DistanceTo calculates the distance between two locations
func DistanceTo(latitude, longitude, distToLatidude, distToLongitude float64) float64 {
	//convert to rad
	latitude = latitude * math.Pi / 180
	longitude = longitude * math.Pi / 180
	distToLatidude = distToLatidude * math.Pi / 180
	distToLongitude = distToLongitude * math.Pi / 180

	//dist = arccos(sin(lat1) · sin(lat2) + cos(lat1) · cos(lat2) · cos(lon1 - lon2)) · R with R radius of the Earth
	return math.Acos(math.Sin(latitude)*math.Sin(distToLatidude)+math.Cos(latitude)*math.Cos(distToLatidude)*math.Cos(longitude-distToLongitude)) * earthRadius
}

//BoundingCoordinates computes the bounding coordinates from a location
func BoundingCoordinates(latitude, longitude, distance float64) []Geolocation {
	//distance in radians on a great circle
	radDist := distance / earthRadius
	radLat := latitude * math.Pi / 180
	radLon := longitude * math.Pi / 180

	minLat := radLat - radDist
	maxLat := radLat + radDist

	var minLon, maxLon float64
	if minLat > -math.Pi/2 && maxLat < math.Pi/2 {
		deltaLon := math.Asin(math.Sin(radDist) / math.Cos(radLat))
		minLon = radLon - deltaLon
		if minLon < -math.Pi {
			minLon += 2 * math.Pi
		}
		maxLon = radLon + deltaLon
		if maxLon > math.Pi {
			maxLon -= 2 * math.Pi
		}
	} else {
		minLat = math.Max(minLat, -math.Pi/2)
		maxLat = math.Min(maxLat, math.Pi/2)
		minLon = -math.Pi
		maxLon = math.Pi
	}

	return []Geolocation{
		{
			Latitude:  minLat * 180 / math.Pi,
			Longitude: minLon * 180 / math.Pi,
		},
		{
			Latitude:  maxLat * 180 / math.Pi,
			Longitude: maxLon * 180 / math.Pi,
		},
	}
}
