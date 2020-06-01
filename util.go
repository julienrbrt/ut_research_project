package main

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

//WriteCSV writes CSV from dataframe
func WriteCSV(df dataframe.DataFrame, path string) error {
	log.Println("Writing CSV...")

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
