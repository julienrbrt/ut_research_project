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
	Ingredients     []string
	IngredientsOnly []string
	Instructions    []string
	Tags            []string
	CookTime        int `json:"cookTime"`
	OvenTime        int `json:"ovenTime"`
	WaitTime        int `json:"waitTime"`
	Rating          struct {
		AverageRating   int `json:"averageRating"`
		NumberOfRatings int `json:"numberOfRatings"`
	} `json:"rating"`
	Nutritions struct {
		SATURATEDFAT struct {
			Name  string  `json:"name"`
			Unit  string  `json:"unit"`
			Value float32 `json:"value"`
		} `json:"SATURATED_FAT"`
		ENERGY struct {
			Name  string  `json:"name"`
			Unit  string  `json:"unit"`
			Value float32 `json:"value"`
		} `json:"ENERGY"`
		PROTEIN struct {
			Name  string  `json:"name"`
			Unit  string  `json:"unit"`
			Value float32 `json:"value"`
		} `json:"PROTEIN"`
		FAT struct {
			Name  string  `json:"name"`
			Unit  string  `json:"unit"`
			Value float32 `json:"value"`
		} `json:"FAT"`
		CARBOHYDRATES struct {
			Name  string  `json:"name"`
			Unit  string  `json:"unit"`
			Value float32 `json:"value"`
		} `json:"CARBOHYDRATES"`
	} `json:"nutritions"`
	ImageURL string
	URL      string `json:"href"`
}

//AHRecipes contains a recipe list from AH
type AHRecipes struct {
	Recipes []AHRecipe `json:"recipes"`
}

//ScrapeAH scrapes a recipe from Albert Heijn Allerhande website
func (r *AHRecipe) ScrapeAH(recipeURL string) {
	//get url
	r.URL = recipeURL

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

			r.IngredientsOnly = append(r.IngredientsOnly, strings.ToLower(ingredient))
			r.Ingredients = append(r.Ingredients, strings.TrimSpace(i.DOM.Children().Text()))
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
			r.Tags = append(r.Tags, strings.ToLower(strings.TrimSpace(i.Text)))
		})
	})

	//get image
	c.OnHTML("li.responsive-image", func(e *colly.HTMLElement) {
		r.ImageURL, _ = e.DOM.Attr("data-phone-src")
	})

	c.Visit(recipeURL)
}

//ScrapeNAH gets N recipes from AH Allerhande Search API
func ScrapeNAH(n int) (*AHRecipes, error) {
	recipesURL := "https://www.ah.nl/allerhande2/api/recipe-search?searchText=&filters=[%22menugang;hoofdgerecht%22]&size=" + strconv.Itoa(n)

	resp, err := http.Get(recipesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//read json as byte array
	byteValue, _ := ioutil.ReadAll(resp.Body)

	//unmarshal json
	var recipes AHRecipes
	err = json.Unmarshal(byteValue, &recipes)
	if err != nil {
		return nil, err
	}

	for i := range recipes.Recipes {
		log.Printf("Getting recipe %d / %d\n", i+1, n)
		recipes.Recipes[i].ScrapeAH("https://www.ah.nl" + recipes.Recipes[i].URL)
	}

	return &recipes, nil
}
