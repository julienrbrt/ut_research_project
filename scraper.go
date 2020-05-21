package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

//Recipe contains a recipe data
type Recipe struct {
	Title        string
	TotalTime    int // time in minutes
	Ingredients  map[string]string
	Instructions []string
	Tags         []string
	ImageURL     string
	URL          string
}

func scrapeAH(recipeURL string) *Recipe {
	//get url
	recipe := Recipe{URL: recipeURL, Ingredients: make(map[string]string)}

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
		recipe.Title = e.Text
	})

	//get cooking time
	c.OnHTML("li.cooking-time", func(e *colly.HTMLElement) {
		//if totalTime greater than 0 then we've already parsed it. Probably duplicate cooking time in the html
		if recipe.TotalTime > 0 {
			return
		}

		//get number regex
		re := regexp.MustCompile("[0-9]+")

		e.ForEach("li", func(_ int, t *colly.HTMLElement) {
			time, err := strconv.Atoi(re.FindString(t.Text))
			if err != nil {
				return
			}

			if strings.Contains(t.Text, "uur") {
				time *= 60
			}

			recipe.TotalTime += time
		})
	})

	//get ingredients
	c.OnHTML("li[itemprop=\"ingredients\"]", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, i *colly.HTMLElement) {
			ingredient, _ := i.DOM.Attr("data-description-singular")
			recipe.Ingredients[strings.TrimSpace(ingredient)] = strings.TrimSpace(i.DOM.Children().Text())
		})
	})

	//get instructions
	c.OnHTML("section[itemprop=\"recipeInstructions\"]", func(e *colly.HTMLElement) {
		e.ForEach("li", func(_ int, i *colly.HTMLElement) {
			recipe.Instructions = append(recipe.Instructions, i.Text)
		})
	})

	//get tags
	c.OnHTML("section.tags", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, i *colly.HTMLElement) {
			recipe.Tags = append(recipe.Tags, strings.TrimSpace(i.Text))
		})
	})

	//get image
	c.OnHTML("li.responsive-image", func(e *colly.HTMLElement) {
		recipe.ImageURL, _ = e.DOM.Attr("data-phone-src")
	})

	c.Visit(recipeURL)

	return &recipe
}
