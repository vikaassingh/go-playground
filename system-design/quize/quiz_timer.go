package main

import "fmt"

// Question(map[Question]Answer{}), Answer, Score

var (
	questions = map[string]string{
		"What is the capital of France?":                  "Paris",
		"What is 2 + 2?":                                  "4",
		"What is the largest planet in our solar system?": "Jupiter",
	}

	score = 0
)

func main() {
	StartQuiz()
	fmt.Printf("Your final score is: %d out of %d\n", score, len(questions))
}

func StartQuiz() {
	for q, a := range questions {
		fmt.Println(q)
		answer := ReadFromTerminal()
		if answer == a {
			score++
		}
	}
}

func ReadFromTerminal() string {
	var input string
	fmt.Scanln(&input)
	return input
}
