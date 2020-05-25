package main

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
