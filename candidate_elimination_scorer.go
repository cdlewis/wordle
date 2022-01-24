package main

import (
	"strings"
)

type Position struct {
	Letter   byte
	Position int
}

type EliminationScorer struct {
	totalWords                      int
	wordsWithoutLetter              map[byte][]string
	wordsWithLetterAtPosition       map[byte]map[int][]string
	wordsWithLetterButNotAtPosition map[byte]map[int][]string
}

func NewEliminationScorer(possibleAnswers []string, possibleGuesses []string) *EliminationScorer {

	// Pre-process the input words for quick lookup by letter with/without arbitrary position

	wordsWithLetter := map[byte][]string{}
	wordsWithoutLetter := map[byte][]string{}
	wordsWithLetterAtPosition := map[byte]map[int][]string{}
	wordsWithLetterButNotAtPosition := map[byte]map[int][]string{}

	for _, word := range possibleAnswers {
		for index, letter := range word {
			if _, ok := wordsWithLetter[byte(letter)]; !ok {
				wordsWithLetter[byte(letter)] = []string{}
			}

			if _, ok := wordsWithLetterAtPosition[byte(letter)]; !ok {
				wordsWithLetterAtPosition[byte(letter)] = map[int][]string{}
			}

			if _, ok := wordsWithLetterAtPosition[byte(letter)][index]; !ok {
				wordsWithLetterAtPosition[byte(letter)][index] = []string{}
			}

			if _, ok := wordsWithLetterButNotAtPosition[byte(letter)]; !ok {
				wordsWithLetterButNotAtPosition[byte(letter)] = map[int][]string{}
			}

			wordsWithLetter[byte(letter)] = append(wordsWithLetter[byte(letter)], word)

			wordsWithLetterAtPosition[byte(letter)][index] = append(wordsWithLetterAtPosition[byte(letter)][index], word)

			for i := 0; i < 5; i++ {
				if _, ok := wordsWithLetterButNotAtPosition[byte(letter)][index]; !ok {
					wordsWithLetterButNotAtPosition[byte(letter)][index] = []string{}
				}

				if i == index {
					continue
				}

				wordsWithLetterButNotAtPosition[byte(letter)][i] = append(wordsWithLetterButNotAtPosition[byte(letter)][i], word)
			}
		}
	}
	for _, word := range possibleAnswers {
		hasLetters := map[byte]bool{}
		for _, letter := range word {
			hasLetters[byte(letter)] = true
		}

		for letter, _ := range wordsWithLetter {
			if _, ok := wordsWithoutLetter[byte(letter)]; !ok {
				wordsWithoutLetter[byte(letter)] = []string{}
			}

			if _, ok := hasLetters[letter]; !ok {
				wordsWithoutLetter[byte(letter)] = append(wordsWithoutLetter[byte(letter)], word)
			}
		}
	}

	return &EliminationScorer{
		totalWords:                      len(possibleAnswers),
		wordsWithLetterAtPosition:       wordsWithLetterAtPosition,
		wordsWithoutLetter:              wordsWithoutLetter,
		wordsWithLetterButNotAtPosition: wordsWithLetterButNotAtPosition,
	}
}

func (e *EliminationScorer) score(candidateWord string, targetWord string) int {
	positions := []Position{}
	notAtPosition := []Position{}
	withoutLetter := []byte{}

	for index, letter := range candidateWord {
		isExact := targetWord[index] == candidateWord[index]
		if isExact {
			positions = append(positions, Position{Letter: byte(letter), Position: index})
		}

		// Faster to just exhaustively check
		isWildcard := !isExact && (candidateWord[index] == targetWord[0] ||
			candidateWord[index] == targetWord[1] ||
			candidateWord[index] == targetWord[2] ||
			candidateWord[index] == targetWord[3] ||
			candidateWord[index] == targetWord[4])
		if isWildcard {
			notAtPosition = append(notAtPosition, Position{Letter: byte(candidateWord[index]), Position: index})
		}

		if !isExact && !isWildcard {
			withoutLetter = append(withoutLetter, byte(letter))
		}
	}

	prunedCandidates := [][]string{}
	for _, position := range positions {
		prunedCandidates = append(prunedCandidates, e.wordsWithLetterAtPosition[position.Letter][position.Position])
	}
	for _, position := range notAtPosition {
		prunedCandidates = append(prunedCandidates, e.wordsWithLetterButNotAtPosition[position.Letter][position.Position])
	}
	for _, letter := range withoutLetter {
		prunedCandidates = append(prunedCandidates, e.wordsWithoutLetter[letter])
	}

	eliminatedWords := e.totalWords

	for {
		smallestString := "<>"

		for _, candidateList := range prunedCandidates {
			// If a list is every empty just return -- impossible to find more candidates present in all lists
			if len(candidateList) == 0 {
				return eliminatedWords
			}

			if smallestString == "<>" || strings.Compare(candidateList[0], smallestString) != 1 {
				smallestString = candidateList[0]
			}
		}

		// Check if all lists have our candidate for smallest string
		allListsHaveString := true
		for index, candidateList := range prunedCandidates {
			if candidateList[0] == smallestString {
				// Drop string from the list and continue
				prunedCandidates[index] = prunedCandidates[index][1:]
			} else {
				allListsHaveString = false
				// Keep looking so we can drop it from other lists as part of pruning
			}
		}

		if allListsHaveString {
			eliminatedWords--
		}

		// Kind of confusing but this will break eventually
	}

	return eliminatedWords
}
