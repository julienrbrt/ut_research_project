package util

import (
	"errors"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

//RemoveDuplicatesUnordered removes duplicates and ignores order
func RemoveDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}

	return result
}

//Unique returns a unique slice
func Unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

//WriteCSV writes CSV from dataframe
func WriteCSV(df dataframe.DataFrame, path string) error {
	log.Printf("Writing CSV in %s...\n", path)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := df.WriteCSV(f); err != nil {
		return err
	}

	return nil
}

//LoadCSV loads data from on disk CSV as dataframe
func LoadCSV(path string) dataframe.DataFrame {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	df := dataframe.ReadCSV(strings.NewReader(string(content)),
		dataframe.WithDelimiter(','),
		dataframe.HasHeader(true))

	return df
}

//CosineSimilarity calculates the cosine similarity of two values
func CosineSimilarity(a []float64, b []float64) (cosine float64, err error) {
	count := 0
	lengthA := len(a)
	lengthB := len(b)

	if lengthA > lengthB {
		count = lengthA
	} else {
		count = lengthB
	}

	sumA := 0.0
	s1 := 0.0
	s2 := 0.0

	for k := 0; k < count; k++ {

		if k >= lengthA {
			s2 += b[k] * b[k]
			continue
		}

		if k >= lengthB {
			s1 += a[k] * a[k]
			continue
		}

		sumA += a[k] * b[k]
		s1 += a[k] * a[k]
		s2 += b[k] * b[k]
	}

	if s1 == 0 || s2 == 0 {
		return 0.0, errors.New("Vectors should not be null (all zeros)")
	}

	return sumA / (math.Sqrt(s1) * math.Sqrt(s2)), nil
}

//SS2SF convers a String Slice to a String Float
func SS2SF(values []string) ([]float64, error) {
	slice := []float64{}
	for i := range values {
		f, err := strconv.ParseFloat(values[i], 64)
		if err != nil {
			return nil, err
		}
		slice = append(slice, f)
	}

	return slice, nil
}
