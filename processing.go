package main

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

const recipesCSVPath = "data/item_recipes.csv"

//TransformToCSV transforms data in a csv acceptable format
func (recipes *AHRecipes) TransformToCSV() (*[]string, *[][]string) {
	log.Println("Processing...")

	headers := []string{"id", "title", "totalTime", "averageRating", "numberOfRatings", "imageURL", "URL"}
	records := [][]string{}

	//add all tags and ingredients from recipes
	var tags []string
	var ingredients []string
	for _, recipe := range recipes.Recipes {
		tags = append(tags, recipe.Tags...)
		ingredients = append(ingredients, recipe.IngredientsOnly...)
	}

	//clean ingredients and tags
	tags = CleanIngredientsAndTags(tags, false)
	ingredients = CleanIngredientsAndTags(ingredients, true)

	//append to headers
	headers = append(headers, tags...)
	headers = append(headers, ingredients...)

	for i, recipe := range recipes.Recipes {
		//fill in data
		data := []string{
			strconv.Itoa(i),
			recipe.Title,
			strconv.Itoa(recipe.CookTime + recipe.OvenTime + recipe.WaitTime),
			strconv.Itoa(recipe.Rating.AverageRating),
			strconv.Itoa(recipe.Rating.NumberOfRatings),
			recipe.ImageURL,
			recipe.URL,
		}

		//clean ingredients and tags
		recipe.Tags = CleanIngredientsAndTags(recipe.Tags, false)
		recipe.IngredientsOnly = CleanIngredientsAndTags(recipe.IngredientsOnly, true)

		//map of contained tags
		set := make(map[string]bool)
		for _, t := range recipe.Tags {
			set[t] = true
		}

		//add tags
		for _, t := range tags {
			if set[t] {
				data = append(data, "1")
			} else {
				data = append(data, "0")
			}
		}

		//map of contained ingredients
		set = make(map[string]bool)
		for _, i := range recipe.IngredientsOnly {
			set[i] = true
		}

		//add ingredients
		for _, i := range ingredients {
			if set[i] {
				data = append(data, "1")
			} else {
				data = append(data, "0")
			}
		}

		records = append(records, data)
	}

	return &headers, &records
}

//CleanIngredientsAndTags clean recipes ingredient or tags list
//TODO find a better way for text processing
func CleanIngredientsAndTags(data []string, isIngredient bool) []string {
	for i := range data {
		//remove non-alphanumeric chracter
		reg, err := regexp.Compile("[^A-zÀ-ú ]+")
		if err != nil {
			log.Fatalln(err)
		}
		data[i] = reg.ReplaceAllString(data[i], "")

		//remove AH brands
		data[i] = strings.ReplaceAll(data[i], "ah biologisch", "")
		data[i] = strings.ReplaceAll(data[i], "ah basic", "")

		//remove food storage
		data[i] = strings.ReplaceAll(data[i], "houdbare", "")
		data[i] = strings.ReplaceAll(data[i], "koelverse", "")
		data[i] = strings.ReplaceAll(data[i], "diepvries", "")
		data[i] = strings.ReplaceAll(data[i], "verse", "")
		data[i] = strings.ReplaceAll(data[i], "gemalen", "")
		data[i] = strings.ReplaceAll(data[i], "halfgedroogde ", "")
		data[i] = strings.ReplaceAll(data[i], "gedroogde ", "")
		data[i] = strings.ReplaceAll(data[i], "gedroogd ", "")
		data[i] = strings.ReplaceAll(data[i], "gesneden", "")
		data[i] = strings.ReplaceAll(data[i], "zongerijpte", "")
		data[i] = strings.ReplaceAll(data[i], "stuckjes", "")
		data[i] = strings.ReplaceAll(data[i], "stukjes", "")
		data[i] = strings.ReplaceAll(data[i], "a la minute", "")
		data[i] = strings.ReplaceAll(data[i], "kruimige", "")
		data[i] = strings.ReplaceAll(data[i], "fijne", "")
		data[i] = strings.ReplaceAll(data[i], "fijn", "")
		data[i] = strings.ReplaceAll(data[i], "warmgerookte", "")
		data[i] = strings.ReplaceAll(data[i], "ongezouten", "")
		data[i] = strings.ReplaceAll(data[i], "gezouten", "")
		data[i] = strings.ReplaceAll(data[i], "kleine", "")
		data[i] = strings.ReplaceAll(data[i], "klein", "")
		data[i] = strings.ReplaceAll(data[i], "grote", "")
		data[i] = strings.ReplaceAll(data[i], "groot", "")
		data[i] = strings.ReplaceAll(data[i], "biologische ", "")
		data[i] = strings.ReplaceAll(data[i], "biologisch ", "")
		data[i] = strings.ReplaceAll(data[i], "halfvolle ", "")
		data[i] = strings.ReplaceAll(data[i], "volle ", "")
		data[i] = strings.ReplaceAll(data[i], "mager ", "")
		data[i] = strings.ReplaceAll(data[i], "magere ", "")
		data[i] = strings.ReplaceAll(data[i], "halfomhalf ", "")
		data[i] = strings.ReplaceAll(data[i], "zoete", "")
		data[i] = strings.ReplaceAll(data[i], "iets ", "")
		data[i] = strings.ReplaceAll(data[i], "roerbakgroenten", "roerbakgroente")

		//replace some ingredients
		toReplace := []string{"kruidenmix", "aardappel", "saus", "boter", "olie", "groentemix", "roerbakgroente", "brood"}
		for _, ingredient := range toReplace {
			if strings.Contains(data[i], ingredient) {
				data[i] = ingredient
			}

		}

		//trimm space
		data[i] = strings.TrimSpace(data[i])
		spaceReg := regexp.MustCompile(`\s+`)
		data[i] = spaceReg.ReplaceAllString(data[i], "_")

		//append ingredient or tag name
		if isIngredient {
			data[i] = "ingredient_" + data[i]
		} else {
			data[i] = "tag_" + data[i]
		}

	}

	//remove duplicates
	data = RemoveDuplicatesUnordered(data)

	return data
}

//LoadData recipes data from internet
func LoadData(n int, writeCSV bool) (dataframe.DataFrame, error) {
	//scrape recipes
	recipes, err := ScrapeNAH(n)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	//processing
	header, records := recipes.TransformToCSV()

	if writeCSV {
		log.Println("Writing CSV...")
		if err := WriteCSV(recipesCSVPath, header, records); err != nil {
			return dataframe.DataFrame{}, err
		}
	}

	//load data
	data := [][]string{*header}
	data = append(data, *records...)

	df := dataframe.LoadRecords(data)

	return df, nil
}

//LoadDataFromCSV recipes data from on disk CSV
func LoadDataFromCSV() (dataframe.DataFrame, error) {
	content, err := ioutil.ReadFile(recipesCSVPath)
	if err != nil {
		return dataframe.DataFrame{}, nil
	}

	df := dataframe.ReadCSV(strings.NewReader(string(content)),
		dataframe.WithDelimiter(','),
		dataframe.HasHeader(true))

	return df, nil
}
