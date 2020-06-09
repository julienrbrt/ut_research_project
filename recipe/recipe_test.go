package recipe

import "testing"

//TestScrapeAH tests ScrapeAH
func TestScrapeAH(t *testing.T) {
	expectedRecipe := Recipe{
		Title:    "Pasta pesto vegetarisch",
		Tags:     []string{"snel", "vegetarish", "italians", "wat eten we vandaag", "koken", "5-ingrediënten"},
		ImageURL: "https://static.ah.nl/static/recepten/img_RAM_PRD123716_890x594_JPG.jpg",
		URL:      "https://www.ah.nl/allerhande/recept/R-R1192908/pasta-pesto-vegetarisch",
	}

	//scrape recipe
	recipe := Recipe{}
	recipe.ScrapeAH(expectedRecipe.URL)

	if recipe.Title != expectedRecipe.Title {
		t.Errorf("Title is incorrect, got '%s', want '%s'", recipe.Title, expectedRecipe.Title)
	}

	if recipe.Ingredients == nil || len(recipe.Ingredients) == 0 {
		t.Errorf("Ingredients are incorrect, got '%v', want a positive non nil value", recipe.Ingredients)
	}

	if len(recipe.Ingredients) != len(recipe.IngredientsOnly) {
		t.Errorf("Ingredients are incorrect, got '%d', want '%d'", len(recipe.Ingredients), len(recipe.IngredientsOnly))
	}

	if recipe.Instructions == nil || len(recipe.Instructions) == 0 {
		t.Errorf("Instructions are incorrect, got '%v', want a positive non nil value", recipe.Instructions)
	}

	if len(recipe.Tags) != len(expectedRecipe.Tags) {
		t.Errorf("Tags are incorrect, got '%v', want '%v'", recipe.Tags, expectedRecipe.Tags)
	}

	if recipe.ImageURL != expectedRecipe.ImageURL {
		t.Errorf("ImageURL is incorrect, got '%s', want '%s'", recipe.ImageURL, expectedRecipe.ImageURL)
	}

	if recipe.URL != expectedRecipe.URL {
		t.Errorf("URL is incorrect, got '%s', want '%s'", recipe.URL, expectedRecipe.URL)
	}
}

//TestScrapeNAH tests ScrapeNAH
func TestScrapeNAH(t *testing.T) {
	//scrape 10 recipes from AH
	expectedRecipesLength := 10
	recipes, _ := ScrapeNAH(expectedRecipesLength)

	if len(recipes.Recipes) != 10 {
		t.Errorf("The number of recipes is incorrect, got '%d', want '%d'", len(recipes.Recipes), expectedRecipesLength)
	}
}
