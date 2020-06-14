package util

import "testing"

func TestRemoveDuplicatesUnordered(t *testing.T) {
	input := []string{"fresh", "meat", "vegetarian", "meat", "fresh", "fresh"}
	expectedOutput := []string{"fresh", "meat", "vegetarian"}

	//remove duplicate
	output := RemoveDuplicatesUnordered(input)

	if len(output) != len(expectedOutput) {
		t.Errorf("Output is incorrect, got '%v', want '%v'", output, expectedOutput)
	}
}

func TestCosineSimilarity(t *testing.T) {

	a := []float64{1, 1, 0, 0, 1}
	b := []float64{1, 1, 0, 0, 1}
	cos, err := CosineSimilarity(a, b)
	if err != nil {
		t.Error(err)
	}
	if cos < 0.99 {
		t.Error("Expected similarity of 1, got instead ", cos)
	}
	a = []float64{0, 1, 0, 1, 1}
	b = []float64{1, 0, 1, 0, 0}
	cos, err = CosineSimilarity(a, b)
	if err != nil {
		t.Error(err)
	}
	if cos != 0 {
		t.Error("Expected similarity of 0, got instead ", cos)
	}
	a = []float64{1, 1, 0}
	b = []float64{1, 0, 1}
	cos, err = CosineSimilarity(a, b)
	if err != nil {
		t.Error(err)
	}
	if cos < 0.49999 || cos > 0.5 {
		t.Error("Expected similarity of 0.5, got instead ", cos)
	}
	a = []float64{0, 1, 1, 1, 0}
	b = []float64{1, 0}
	cos, err = CosineSimilarity(a, b)
	if err != nil {
		t.Error(err)
	}
	if cos != 0 {
		t.Error("Expected similarity of 0, got instead ", cos)
	}
}
