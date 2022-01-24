package main

import (
	"fmt"
	"sort"
)

type Constraint interface {
	Priority() int
	Describe() string
	Evaluate(word string) bool
}

func FilterByConstraints(words []string, constraints []Constraint) []string {
	result := []string{}

	for _, word := range words {
		satisfied := true

		for _, constraint := range constraints {
			if !constraint.Evaluate(word) {
				satisfied = false
				break
			}
		}

		if satisfied {
			result = append(result, word)
		}
	}

	return result
}

type WithoutLetterConstraint struct {
	Letter rune
}

func (c WithoutLetterConstraint) Evaluate(word string) bool {
	for _, i := range word {
		if i == c.Letter {
			return false
		}
	}

	return true
}

func (c WithoutLetterConstraint) Priority() int {
	return 2
}

func (c WithoutLetterConstraint) Describe() string {
	return fmt.Sprintf("Without letter %s", string(c.Letter))
}

type WithLetterAtPositionConstraint struct {
	Letter   byte
	Position int
}

func (c WithLetterAtPositionConstraint) Evaluate(word string) bool {
	return word[c.Position] == c.Letter
}

func (c WithLetterAtPositionConstraint) Priority() int {
	return 0
}

func (c WithLetterAtPositionConstraint) Describe() string {
	return fmt.Sprintf("With letter %s at position %v", string(c.Letter), c.Position)
}

type WithLetterNotAtPositionConstraint struct {
	Letter   rune
	Position int
}

func (c WithLetterNotAtPositionConstraint) Evaluate(word string) bool {
	for idx, letter := range word {
		if idx == c.Position {
			continue
		}

		if letter == c.Letter {
			return true
		}
	}

	return false
}

func (c WithLetterNotAtPositionConstraint) Priority() int {
	return 1
}

func (c WithLetterNotAtPositionConstraint) Describe() string {
	return fmt.Sprintf("With letter %s at any position other than %v", string(c.Letter), c.Position)
}

type ConstraintBasedEliminationScorer struct {
	possibleAnswers []string
}

func (e *ConstraintBasedEliminationScorer) score(candidateWord string, targetWord string) int {
	constraints := []Constraint{}

	for index, letter := range candidateWord {
		isExact := targetWord[index] == candidateWord[index]
		if isExact {
			constraints = append(constraints, WithLetterAtPositionConstraint{Letter: byte(letter), Position: index})
		}

		// Faster to just exhaustively check
		isWildcard := !isExact && (candidateWord[index] == targetWord[0] ||
			candidateWord[index] == targetWord[1] ||
			candidateWord[index] == targetWord[2] ||
			candidateWord[index] == targetWord[3] ||
			candidateWord[index] == targetWord[4])
		if isWildcard {
			constraints = append(constraints, WithLetterNotAtPositionConstraint{Letter: rune(candidateWord[index]), Position: index})
		}

		if !isExact && !isWildcard {
			constraints = append(constraints, WithoutLetterConstraint{Letter: letter})
		}
	}

	// order constraints by priority
	sort.Slice(constraints, func(i, j int) bool {
		return constraints[i].Priority() < constraints[j].Priority()
	})

	eliminatedWords := len(e.possibleAnswers)

	for _, word := range e.possibleAnswers {
		satisfiedConstraints := true

		for _, c := range constraints {
			if !c.Evaluate(word) {
				satisfiedConstraints = false
				break
			}
		}

		if satisfiedConstraints {
			eliminatedWords--
		}
	}

	return eliminatedWords
}
