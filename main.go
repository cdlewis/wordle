package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/schollz/progressbar/v3"
)

func getWords(filePath string) []string {
	content, _ := ioutil.ReadFile(filePath)
	var words []string
	_ = json.Unmarshal(content, &words)
	sort.Strings(words)
	return words
}

type Result struct {
	Word string
	p50  int
}

func main() {
	withoutLetters := flag.String("withoutLetters", "", "Letters to exclude, e.g. -withoutLetters abc")
	withLettersAtPosition := flag.String("withLettersAtPosition", "", "Letters that must be in a particular position, e.g. -withLettersAtPosition a=1,b=2")
	withLettersNotAtPosition := flag.String("withLettersNotAtPosition", "", "Letters that must not be in a particular position, e.g. withLettersNotAtPosition a=1,b=2")
	flag.Parse()

	userDefinedConstraints := []Constraint{}

	if withoutLetters != nil {
		for _, letter := range *withoutLetters {
			userDefinedConstraints = append(userDefinedConstraints, WithoutLetterConstraint{Letter: letter})
		}
	}

	if withLettersAtPosition != nil && len(*withLettersAtPosition) > 0 {
		for _, pair := range strings.Split(*withLettersAtPosition, ",") {
			letter := pair[0]
			position, _ := strconv.ParseInt(string(pair[2]), 10, 64)

			userDefinedConstraints = append(userDefinedConstraints, WithLetterAtPositionConstraint{
				Letter:   letter,
				Position: int(position),
			})
		}
	}

	if withLettersNotAtPosition != nil && len(*withLettersNotAtPosition) > 0 {
		for _, pair := range strings.Split(*withLettersNotAtPosition, ",") {
			letter := pair[0]
			position, _ := strconv.ParseInt(string(pair[2]), 10, 64)

			userDefinedConstraints = append(userDefinedConstraints, WithLetterNotAtPositionConstraint{
				Letter:   rune(letter),
				Position: int(position),
			})
		}
	}

	if len(userDefinedConstraints) > 0 {
		fmt.Print("Running Wordle solver with the following user-defined constraints:\n\n")
		for _, c := range userDefinedConstraints {
			fmt.Printf("\t * %v\n", c.Describe())
		}
	}
	fmt.Print("\n")

	possibleAnswers := FilterByConstraints(getWords("possible_answers.json"), userDefinedConstraints)
	dictionary := FilterByConstraints(append(getWords("possible_guesses.json"), possibleAnswers...), userDefinedConstraints)
	scorer := ConstraintBasedEliminationScorer{possibleAnswers: possibleAnswers}

	// Worker takes a collection of words and scores them based on their ability to narrow the result set

	worker := func(id int, allocatedWords []string, reportResult func(Result), done func()) {
		defer done()

		for _, startWord := range allocatedWords {
			scores := []int{}

			for _, targetWord := range possibleAnswers {
				if startWord == targetWord {
					continue
				}

				scores = append(scores, scorer.score(startWord, targetWord))
			}

			sort.Ints(scores)

			p50 := scores[len(scores)/2]

			reportResult(Result{Word: startWord, p50: p50})
		}
	}

	// Schedule workers and track the output scores

	bar := progressbar.Default(int64(len(dictionary)), "Scoring candidates")

	resultChannel := make(chan Result, len(dictionary))
	var wg sync.WaitGroup
	chunkSize := len(dictionary) / runtime.NumCPU()
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)

		start := i * chunkSize
		end := i*chunkSize + chunkSize

		allocatedCandidates := dictionary[start:end]

		go worker(i, allocatedCandidates, func(r Result) { bar.Add(1); resultChannel <- r }, func() { wg.Done() })
	}

	wg.Wait()
	close(resultChannel)

	// Sort the results and output the most promising candidates

	results := []Result{}
	for i := range resultChannel {
		results = append(results, i)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].p50 >= results[j].p50
	})

	possibleAnswersMap := map[string]bool{}
	for _, w := range possibleAnswers {
		possibleAnswersMap[w] = true
	}

	fmt.Print("\n\nTop 10 candidates:\n\n") // progress bar doesn't add a new line
	for i, r := range results {
		if i > 10 {
			break
		}

		validAnswerTag := ""
		if _, ok := possibleAnswersMap[r.Word]; ok {
			validAnswerTag = "(valid)"
		}

		fmt.Println("\t *", r.Word, r.p50, validAnswerTag)
	}
}
