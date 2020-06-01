package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
)

const recipesCSVPath = "data/item_recipes.csv"

//asRecipesDf converts a list of recipes as a dataframe
func (recipes *AHRecipes) asRecipesDf() (dataframe.DataFrame, error) {
	log.Println("Processing...")

	var err error
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
	tags, err = cleanIngredientsAndTags(tags, false)
	if err != nil {
		return dataframe.DataFrame{}, err
	}
	ingredients, err = cleanIngredientsAndTags(ingredients, true)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

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
		recipe.Tags, err = cleanIngredientsAndTags(recipe.Tags, false)
		if err != nil {
			return dataframe.DataFrame{}, nil
		}
		recipe.IngredientsOnly, err = cleanIngredientsAndTags(recipe.IngredientsOnly, true)
		if err != nil {
			return dataframe.DataFrame{}, nil
		}

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

	//load as dataframe
	df := dataframe.LoadRecords(append([][]string{headers}, records...))

	return df, nil
}

//cleanIngredientsAndTags clean recipes ingredient or tags list
//TODO find a better way for text processing
func cleanIngredientsAndTags(data []string, isIngredient bool) ([]string, error) {
	for i := range data {
		//remove non-alphanumeric chracter
		reg, err := regexp.Compile("[^A-zÀ-ú ]+")
		if err != nil {
			return []string{}, nil
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

	return data, nil
}

//LoadRecipesData of N recipes from internet
func LoadRecipesData(n int, writeCSV bool) dataframe.DataFrame {
	//scrape recipes
	recipes, err := scrapeNAH(n)
	if err != nil {
		log.Fatalln(err)
	}

	//processing and load data
	df, err := recipes.asRecipesDf()
	if err != nil {
		log.Fatalln(err)
	}

	if writeCSV {
		//save order csv
		if err := WriteCSV(df, recipesCSVPath); err != nil {
			log.Fatalln(err)
		}
	}

	return df
}
