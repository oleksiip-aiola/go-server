package search

import (
	"strings"
	"unicode"

	snowballeng "github.com/kljensen/snowball/english"
)

func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	tokens = stopWordFilter(tokens)
	tokens = stemmerFilter(tokens)

	return tokens
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))

	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func stopWordFilter(tokens []string) []string {
	var stopwords = map[string]struct{}{
		"a":    {},
		"and":  {},
		"be":   {},
		"have": {},
		"i":    {},
		"in":   {},
		"of":   {},
		"that": {},
		"the":  {},
		"to":   {},
		"with": {},
	}

	r := make([]string, 0, len(tokens))

	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			r = append(r, token)
		}
	}

	return r
}

func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))

	for i, token := range tokens {
		r[i] = snowballeng.Stem(token, false)
	}

	return r
}
