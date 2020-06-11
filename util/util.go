package util

import (
	"io/ioutil"
	"log"
	"os"
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
