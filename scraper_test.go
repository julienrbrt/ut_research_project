package main

import "testing"

//TestScrapeAH tests scrapeAH
func TestScrapeAH(t *testing.T) {
	expectedOutput := Recipe{
		Title:     "Pasta pesto vegetarisch",
		TotalTime: 15,
		Tags:      []string{"snel", "vegetarish", "italians", "wat eten we vandaag", "koken", "5-ingrediënten"},
		ImageURL:  "https://static.ah.nl/static/recepten/img_RAM_PRD123716_890x594_JPG.jpg",
		URL:       "https://www.ah.nl/allerhande/recept/R-R1192908/pasta-pesto-vegetarisch",
	}

	//scrape recipe
	output := scrapeAH(expectedOutput.URL)

	if output.Title != expectedOutput.Title {
		t.Errorf("Title is incorrect, got '%s', want '%s'", output.Title, expectedOutput.Title)
	}

	if output.TotalTime != expectedOutput.TotalTime {
		t.Errorf("TotalTime is incorrect, got '%d', want '%d'", output.TotalTime, expectedOutput.TotalTime)
	}

	if output.Ingredients == nil || len(output.Ingredients) == 0 {
		t.Errorf("Ingredients are incorrect, got '%v', want a positive non nil value", output.Ingredients)
	}

	if output.Instructions == nil || len(output.Instructions) == 0 {
		t.Errorf("Instructions are incorrect, got '%v', want a positive non nil value", output.Instructions)
	}

	if len(output.Tags) != len(expectedOutput.Tags) {
		t.Errorf("Tags are incorrect, got '%v', want '%v'", output.Tags, expectedOutput.Tags)
	}

	if output.ImageURL != expectedOutput.ImageURL {
		t.Errorf("ImageURL is incorrect, got '%s', want '%s'", output.ImageURL, expectedOutput.ImageURL)
	}

	if output.URL != expectedOutput.URL {
		t.Errorf("URL is incorrect, got '%s', want '%s'", output.URL, expectedOutput.URL)
	}
}
