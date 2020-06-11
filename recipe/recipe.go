package recipe

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/gocolly/colly/v2"
	"github.com/julienrbrt/ut_research_project/util"
)

//Recipe contains a recipe data from AH Allerhande website
type Recipe struct {
	Title           string `json:"title"`
	Ingredients     []string
	IngredientsOnly []string
	Instructions    []string
	Tags            []string
	CookTime        int `json:"cookTime"`
	OvenTime        int `json:"ovenTime"`
	WaitTime        int `json:"waitTime"`
	ImageURL        string
	URL             string `json:"href"`
}

//Recipes contains a recipe list from AH
type Recipes struct {
	Recipes []Recipe `json:"recipes"`
}

//ScrapeAH scrapes a recipe from Albert Heijn Allerhande website
func (r *Recipe) ScrapeAH(recipeURL string) {
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
func ScrapeNAH(n int) (*Recipes, error) {
	recipesURL := "https://www.ah.nl/allerhande2/api/recipe-search?searchText=&filters=[%22menugang;hoofdgerecht%22]&size=" + strconv.Itoa(n)

	resp, err := http.Get(recipesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//read json as byte array
	byteValue, _ := ioutil.ReadAll(resp.Body)

	//unmarshal json
	var recipes Recipes
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

//transformToDF converts a list of recipes as a dataframe
func (recipes *Recipes) transformToDF() (dataframe.DataFrame, error) {
	log.Println("Processing...")

	var err error
	headers := []string{"id", "title", "totalTime", "imageURL", "URL"}
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
			strconv.Itoa(i + 1),
			recipe.Title,
			strconv.Itoa(recipe.CookTime + recipe.OvenTime + recipe.WaitTime),
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
	data = util.RemoveDuplicatesUnordered(data)

	return data, nil
}

//RecipesData of N recipes from internet
func RecipesData(n int, csvPath string) (dataframe.DataFrame, error) {
	//scrape recipes
	recipes, err := ScrapeNAH(n)
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	//processing and load data
	df, err := recipes.transformToDF()
	if err != nil {
		return dataframe.DataFrame{}, err
	}

	if csvPath != "" {
		//save order csv
		if err := util.WriteCSV(df, csvPath); err != nil {
			return dataframe.DataFrame{}, err
		}
	}

	return df, nil
}
