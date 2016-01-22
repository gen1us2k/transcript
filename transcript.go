package transcript

import (
	"regexp"
	"os"
	"log"
	"bufio"
	"strings"
	"fmt"
)

var cmudict map[string]string

func GetTranscription(s string) string {
	words := MatchAllWords(s)
	convString := ""
	for _, word := range words {
		up := strings.ToUpper(word)
		isUpper := word != strings.ToLower(word)
		pron := cmudict[up]
		pron = strings.ToLower(pron)
		if pron == "" {
			fmt.Println(up)
		} else {
			phonemes := GetPhonemes(pron, true)
			fixed := FixPhonemes(phonemes)
			for i, phoneme := range fixed {
				bare := StripAccent(phoneme)
				d := strings.ToLower(bare)
				spelling := EnSpellings[d]
				out := phoneme

				if spelling != "" {
					out = spelling
				}

				// Capitalize the first letter of output if the original word was not
				// lowercase.
				if i == 0 && isUpper {
					cap := strings.ToLower(out[0:1])
					if len(out) > 1 {
						cap += out[1:]
					}
					out = cap
				}
				convString += out
			}
		}
		convString += " "

	}
	return convString
}

func TransliterateRussian(s string) string {
	transcripted := GetTranscription(s)
	transliterated := transcripted
	for k, v := range RussianSpellings {
		r, _ := regexp.Compile(k)
		transliterated = r.ReplaceAllString(transliterated, v)
	}
	return transliterated
}

func LoadDict(filename string){
	cmudict = make(map[string]string)
	cmufile, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer cmufile.Close()

	scanner := bufio.NewScanner(cmufile)
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		// Skip comments
		if line[:3] == ";;;" {
			continue
		}
		flds := strings.Split(line, "  ")
		word := flds[0]
		pron := flds[1]
		cmudict[word] = pron
	}
}

func StripIndex(word string) string {
	n := len(word)
	if n > 3 {
		last := word[n-1]
		if last == ')' {
			next := word[n-3]
			if next == '(' {
				word = word[:n-3]
			}
		}
	}
	return word
}

func GetPhonemes(pron string, accent bool) []string {
	phonemes := strings.Split(pron, " ")
	if !accent {
		for i, phoneme := range phonemes {
			phonemes[i] = StripAccent(phoneme)
		}
	}
	return phonemes
}

func StripAccent(phoneme string) string {
	n := len(phoneme)
	last := phoneme[n-1]
	if last == '0' || last == '1' || last == '2' {
		phoneme = phoneme[:n-1]
	}
	return phoneme
}

var AllWordRegexp = regexp.MustCompile(`\pL+(\pP+\pL+)*`)


// MatchAllWords matches the words in a text, including any embedded apostrophes.
// It also matches wordlike expressions that contain nonword characters, such as email or web addresses.
func MatchAllWords(text string) []string {
	return AllWordRegexp.FindAllString(text, -1)
}

func FixPhonemes(phonemes []string) []string {
	n := len(phonemes)
	newPhonemes := make([]string, 0, 2*n)
	for i := 0; i < n; i++ {
		phoneme := phonemes[i]
		if len(phoneme) > 2 && phoneme[:2] == "ER" {
			// Use ahx r instead of erx.
			newPhonemes = append(newPhonemes, "AH"+string(phoneme[2]))
			phoneme = "R"
		} else if i < n-1 {
			if phoneme == "HH" && phonemes[i+1] == "W" {
				// Use wh instead of hw.
				phoneme = "WH"
				i++
			}
		}
		newPhonemes = append(newPhonemes, phoneme)
	}
	n = len(newPhonemes)
	out := make([]string, 0, 2*n)
	for i, ph := range newPhonemes {
		out = append(out, ph)
		if i < n-1 {
			// Use an apostrophe to split up ambiguous combinations of sounds.
			split := false
			if newPhonemes[i+1] == "HH" {
				if ph == "D" || ph == "S" || ph == "T" || ph == "W" || ph == "Z" {
					split = true
				}
			} else if ph == "N" && newPhonemes[i+1] == "G" {
				split = true
			} else if IsVowel(ph) && IsVowel(newPhonemes[i+1]) {
				split = true
			}
			if split {
				out = append(out, "'")
			}
		}
	}
	return out
}

func IsVowel(phoneme string) bool {
	first := phoneme[0]
	return first == 'A' || first == 'E' || first == 'I' || first == 'O' || first == 'U'
}