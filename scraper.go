package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

//AHRecipe contains a recipe data from AH Allerhande website
type AHRecipe struct {
	Title           string `json:"title"`
	Ingredients     map[string]string
	Instructions    []string
	Tags            []string
	Description     string        `json:"description"`
	Classifications []interface{} `json:"classifications"`
	CookTime        int           `json:"cookTime"`
	OvenTime        int           `json:"ovenTime"`
	WaitTime        int           `json:"waitTime"`
	Rating          struct {
		AverageRating   int `json:"averageRating"`
		NumberOfRatings int `json:"numberOfRatings"`
	} `json:"rating"`
	Nutritions struct {
		SATURATEDFAT struct {
			Name  string `json:"name"`
			Unit  string `json:"unit"`
			Value int    `json:"value"`
		} `json:"SATURATED_FAT"`
		ENERGY struct {
			Name  string `json:"name"`
			Unit  string `json:"unit"`
			Value int    `json:"value"`
		} `json:"ENERGY"`
		PROTEIN struct {
			Name  string `json:"name"`
			Unit  string `json:"unit"`
			Value int    `json:"value"`
		} `json:"PROTEIN"`
		FAT struct {
			Name  string `json:"name"`
			Unit  string `json:"unit"`
			Value int    `json:"value"`
		} `json:"FAT"`
		CARBOHYDRATES struct {
			Name  string `json:"name"`
			Unit  string `json:"unit"`
			Value int    `json:"value"`
		} `json:"CARBOHYDRATES"`
	} `json:"nutritions"`
	ImageURL string
	URL      string `json:"href"`
}

//AHRecipeMetadata contains recipes metadata from AH recipes
type AHRecipeMetadata struct {
	Recipes []AHRecipe `json:"recipes"`
}

//scrapeAH scrapes a recipe from Albert Heijn Allerhande website
func (r *AHRecipe) scrapeAH(recipeURL string) {
	//get url
	r.URL = recipeURL
	r.Ingredients = make(map[string]string)

	c := colly.NewCollector(
		// Visit only domains: www.ah.nl
		colly.AllowedDomains("www.ah.nl"),
	)

	//before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	//get title
	c.OnHTML("h1.title.hidden-phones", func(e *colly.HTMLElement) {
		r.Title = e.Text
	})

	//get ingredients
	c.OnHTML("li[itemprop=\"ingredients\"]", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, i *colly.HTMLElement) {
			ingredient, _ := i.DOM.Attr("data-description-singular")
			r.Ingredients[strings.TrimSpace(ingredient)] = strings.TrimSpace(i.DOM.Children().Text())
		})
	})

	//get instructions
	c.OnHTML("section[itemprop=\"recipeInstructions\"]", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, i *colly.HTMLElement) {
			r.Instructions = append(r.Instructions, i.Text)
		})
	})

	//get tags
	c.OnHTML("section.tags", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, i *colly.HTMLElement) {
			r.Tags = append(r.Tags, strings.TrimSpace(i.Text))
		})
	})

	//get image
	c.OnHTML("li.responsive-image", func(e *colly.HTMLElement) {
		r.ImageURL, _ = e.DOM.Attr("data-phone-src")
	})

	c.Visit(recipeURL)
}

//scrapeXAH gets X recipes from AH Allerhande Search API
func scrapeXAH(x int) *[]AHRecipe {
	recipesURL := "https://www.ah.nl/allerhande2/api/recipe-search?searchText=&filters=[%22menugang;hoofdgerecht%22,%22momenten;wat-eten-we-vandaag%22]&size=" + strconv.Itoa(x)

	resp, err := http.Get(recipesURL)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	//read json as byte array
	byteValue, _ := ioutil.ReadAll(resp.Body)

	//unmarshal json
	var metadata AHRecipeMetadata
	err = json.Unmarshal(byteValue, &metadata)
	if err != nil {
		log.Fatalln(err)
	}

	var recipes []AHRecipe
	for _, recipe := range metadata.Recipes {
		recipe.scrapeAH("https://www.ah.nl" + recipe.URL)
		recipes = append(recipes, recipe)
	}

	return &recipes
}
