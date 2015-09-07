package punkt

import (
	"regexp"
	"strings"
)

type PairToken struct {
	First, Second string
}

var ReMultiCharPunct string = `(?:\-{2,}|\.{2,}|(?:\.\s){2,}\.)|(\.\S)`
var endPuncts = []string{ /*`."`, `.'`, `.”`,*/ ":", ",", "?", `?"`, ".)"}

func WordTokenizer(text string) []*PairToken {
	words := strings.Fields(text)
	tokens := make([]*PairToken, 0, len(words))

	multi := regexp.MustCompile(ReMultiCharPunct)

	for _, word := range words {
		// Skip one letter words
		if len(word) == 1 {
			continue
		}

		chars := []rune(word)

		first := word
		second := ""
		for _, punct := range endPuncts {
			if strings.HasSuffix(word, punct) {
				if len(punct) > 1 {
					first = string(chars[:len(chars)-2])
					second = string(chars[len(chars)-2:])
				} else {
					first = string(chars[:len(chars)-1])
					second = string(chars[len(chars)-1:])
				}
			}
		}

		multipunct := multi.FindStringIndex(word)
		if multipunct != nil {
			if strings.HasSuffix(word, ".") && (multipunct[1] != len(word) || multipunct[0]+multipunct[1] == len(word)) {
				first = word[:len(chars)-1]
				second = "."
			} else {
				if multipunct[1] == len(word) {
					first = word[:multipunct[0]]
					second = word[multipunct[0]:]
				} else {
					first = word[:multipunct[1]]
					second = word[multipunct[1]:]
				}
			}
		}

		token := &PairToken{first, second}
		tokens = append(tokens, token)
	}

	return tokens
}
