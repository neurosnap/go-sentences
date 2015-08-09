package punkt

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

var ReNonPunct = regexp.MustCompile(`[^\W\d]`)

// Format of a regular expression to split punctuation from words, excluding period.
const wordTokenFmt string = `(
{{.MultiChar}}
|
(?={{.WordStart}})\S+?				  # Accept word characters until end is found
(?=									  # Sequences marking a word's end
\s|                                   # White-space
$|									  # End-of-string
{{.NonWord}}|{{.MultiChar}}|          # Punctuation
,(?=$|\s|{{.NonWord}}|{{.MultiChar}}) # Comma if at end of word
)
|
\S
)`

type wordTokenStruct struct {
	MultiChar, WordStart, NonWord string
}

/*
Format of a regular expression to find contexts including possible
sentence boundaries. Matches token which the possible sentence boundary
ends, and matches the following token within a lookahead expression
*/
const periodContextFmt string = `
\S*                         # some word material
{{.SentEndChars}}           # a potential sentence ending
(?=(?P<after_tok>
{{.NonWord}}                # either other punctuation
|
\s+(?P<next_tok>\S+)        # or whitespace and some other token
))`

type periodContextStruct struct {
	SentEndChars []byte
	NonWord      string
}

type PunktLanguageVars struct {
	sentEndChars          []byte         // Characters that are candidates for sentence boundaries
	internalPunctuation   string         // Sentence internal punctuation, which indicates an abbreviation if preceded by a period-final token
	reBoundaryRealignment *regexp.Regexp // Used to realign punctuation that should be included in a sentence although it follows the period (or ?, !)
	reWordStart           string         // Excludes some characters from starting word tokens
	reNonWordChars        string         // Characters that cannot appear within words
	reMultiCharPunct      string         // Hyphen and ellipsis are multi-character punctuation
	wordTokenizeFmt       string
	periodContextFmt      string
}

func NewPunktLanguageVars() *PunktLanguageVars {
	return &PunktLanguageVars{
		sentEndChars:        []byte{'.', '?', '!'},
		internalPunctuation: ",:;",
		//reBoundaryRealignment: regexp.MustCompile(`["')\]}]+?(?:\s+|(?=--)|$)`),
		reWordStart:      "[^\\(\"\\`{\\[:;&\\#\\*@\\)}\\]\\-,]",
		reNonWordChars:   `(?:[?!)\";}\]\*:@\'\({\[])`,
		reMultiCharPunct: `(?:\-{2,}|\.{2,}|(?:\.\s){2,}\.)`,
		//wordTokenizeFmt:  wordTokenFmt,
	}
}

// Compile word tokenizer regexp
/*func (p *PunktLanguageVars) ReWordTokenizer() *regexp.Regexp {
	t := template.Must(template.New("wordTokenizer").Parse(p.wordTokenizeFmt))
	var r bytes.Buffer

	t.Execute(&r, wordTokenStruct{
		MultiChar: p.reMultiCharPunct,
		WordStart: p.reWordStart,
		NonWord:   p.reNonWordChars,
	})

	return regexp.MustCompile(r.String())
}*/

// Tokenize a string to split off punctuation other than periods
/*func (p *PunktLanguageVars) WordTokenize(s string) []string {
	return p.ReWordTokenizer().FindAllString(s, -1)
}*/

type WordToken struct {
	First, Second string
}

func (p *PunktLanguageVars) WordTokenizer(text string) []*WordToken {
	words := strings.Fields(text)
	tokens := make([]*WordToken, 0, len(words))

	multi := regexp.MustCompile(p.reMultiCharPunct)
	nonword := regexp.MustCompile(strings.Join([]string{p.reNonWordChars, p.reMultiCharPunct}, "|"))
	//wstart := regexp.MustCompile(p.reNonWordChars)

	for _, word := range words {
		// Skip one letter words
		if len(word) == 1 {
			continue
		}

		first := ""
		second := ""

		if first == "" {
			first = string(word[:1])
			second = string(word[1:])
		}

		punctInWord := nonword.FindStringIndex(word)
		if punctInWord != nil {
			first = word[:punctInWord[0]]
			second = word[punctInWord[0]:]
		}

		if strings.HasSuffix(word, ",") {
			first = word[:len(word)-1]
			second = word[len(word)-1:]
		}

		multipunct := multi.FindStringIndex(word)
		if multipunct != nil {
			if strings.HasSuffix(word, ".") && (multipunct[1] != len(word) || multipunct[0]+multipunct[1] == len(word)) {
				first = word[:len(word)-1]
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

		token := &WordToken{first, second}
		tokens = append(tokens, token)
	}

	return tokens
}

// Compile period context regexp
func (p *PunktLanguageVars) RePeriodContext() *regexp.Regexp {
	t := template.Must(template.New("periodContext").Parse(p.periodContextFmt))
	var r bytes.Buffer

	t.Execute(&r, periodContextStruct{
		SentEndChars: p.sentEndChars,
		NonWord:      p.reNonWordChars,
	})

	return regexp.MustCompile(r.String())
}

// Compiles and returns a regular expression to find contexts including possible sentence boundaries.
func (p *PunktLanguageVars) PeriodContext(s string) []string {
	return p.RePeriodContext().FindAllString(s, -1)
}

func (p *PunktLanguageVars) ReSentEndChars() string {
	return regexp.QuoteMeta(string(p.sentEndChars))
}