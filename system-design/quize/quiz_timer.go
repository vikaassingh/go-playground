package quize

import (
	"fmt"
	"time"
)

var (
	questions = map[string]string{
		"What is the capital of France?":                  "Paris",
		"What is 2 + 2?":                                  "4",
		"What is the largest planet in our solar system?": "Jupiter",
	}
	answerCh = make(chan string)
	score    = 0
)

func StartQuiz() {
	for q, a := range questions {
		fmt.Printf(q + ": ")
		go ReadFromTerminal()

		select {
		case answer := <-answerCh:
			if answer == a {
				score++
			}
		case <-time.After(5 * time.Second):
			fmt.Println("\nTime's up for this question!")
		}
	}
	fmt.Printf("Your final score is: %d out of %d\n", score, len(questions))
}

func ReadFromTerminal() {
	var input string
	fmt.Scanln(&input)
	answerCh <- input
}
