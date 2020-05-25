package main

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/pkg/errors"
)

//removeDuplicatesUnorderedthat removes duplicates and ignores order
func removeDuplicatesUnordered(elements []string) []string {
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

//writeCSV builds the recipe data CSV
func writeCSV(filename string, header *[]string, records *[][]string) error {
	//create csv
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)

	if err := w.Write(*header); err != nil {
		return errors.Wrap(err, "error writing headers to csv")
	}

	for _, record := range *records {
		if err := w.Write(record); err != nil {
			return errors.Wrap(err, "error writing record to csv")
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}

	return nil
}
