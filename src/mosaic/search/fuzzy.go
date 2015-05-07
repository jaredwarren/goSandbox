package search

import (
	"bufio"
	//"errors"
	"fmt"
	"image/color"
	"index/suffixarray"
	"os"
	"regexp"
	//"sort"
	"strings"
	"sync"
)

type Pair struct {
	str1 ColorWord
	str2 ColorWord
}

type ColorWord struct {
	Id     string
	Colors []color.Color
	Score  int
}

func (cw *ColorWord) Equals(hit ColorWord) bool {
	// TODO: compare color
	return true
}

func MakeColorWordFromImage() {

}

type Potential struct {
	Term   ColorWord
	Score  int
	Leven  int
	Method int // 0 - is word, 1 - suggest maps to input, 2 - input delete maps to dictionary, 3 - input delete maps to suggest
}

type Model struct {
	Data            map[string]ColorWord   `json:"data"`
	Maxcount        int                    `json:"maxcount"`
	Suggest         map[string][]ColorWord `json:"suggest"`
	Depth           int                    `json:"depth"`
	Threshold       int                    `json:"threshold"`
	UseAutocomplete bool                   `json:"autocomplete"`
	SuffDivergence  int                    `json:"-"`
	SuffixArr       *suffixarray.Index     `json:"-"`
	SuffixArrConcat string                 `json:"-"`
	sync.RWMutex
}

// For sorting autocomplete suggestions
// to bias the most popular first
type Autos struct {
	Results []string
	Model   *Model
}

func (a Autos) Len() int      { return len(a.Results) }
func (a Autos) Swap(i, j int) { a.Results[i], a.Results[j] = a.Results[j], a.Results[i] }

// Create and initialise a new model
func NewModel() *Model {
	model := new(Model)
	return model.Init()
}

func (model *Model) Init() *Model {
	model.Data = make(map[string]ColorWord)
	model.Suggest = make(map[string][]ColorWord)
	model.Depth = 2
	model.Threshold = 3          // Setting this to 1 is most accurate, but "1" is 5x more memory and 30x slower processing than "4". This is a big performance tuning knob
	model.UseAutocomplete = true // Default is to include Autocomplete
	return model
}

// Change the default depth value of the model. This sets how many
// character differences are indexed. The default is 2.
func (model *Model) SetDepth(val int) {
	model.Lock()
	model.Depth = val
	model.Unlock()
}

// Change the default threshold of the model. This is how many times
// a term must be seen before suggestions are created for it
func (model *Model) SetThreshold(val int) {
	model.Lock()
	model.Threshold = val
	model.Unlock()
}

// Calculate the Levenshtein distance between two strings
func Levenshtein(a, b *ColorWord) int {

	/*la := len(*a)
	lb := len(*b)
	d := make([]int, la+1)
	var lastdiag, olddiag, temp int

	for i := 1; i <= la; i++ {
		d[i] = i
	}
	for i := 1; i <= lb; i++ {
		d[0] = i
		lastdiag = i - 1
		for j := 1; j <= la; j++ {
			olddiag = d[j]
			min := d[j] + 1
			if (d[j-1] + 1) < min {
				min = d[j-1] + 1
			}
			if (*a)[j-1] == (*b)[i-1] {
				temp = 0
			} else {
				temp = 1
			}
			if (lastdiag + temp) < min {
				min = lastdiag + temp
			}
			d[j] = min
			lastdiag = olddiag
		}
	}
	return d[la]*/
	return 0
}

// Add an array of words to train the model in bulk
func (model *Model) Train(terms []ColorWord) {
	for _, term := range terms {
		model.TrainWord(term)
	}
	model.updateSuffixArr()
}

// Train the model word by word
func (model *Model) TrainWord(term ColorWord) {
	model.Lock()
	currentTerm, ok := model.Data[term.Id]
	if ok {
		currentTerm.Score++
	} else {
		term.Score = 0
		model.Data[term.Id] = term
		currentTerm = term
	}
	// Set the max
	if currentTerm.Score > model.Maxcount {
		model.Maxcount = currentTerm.Score
		model.SuffDivergence++
	}
	// If threshold is triggered, store delete suggestion keys
	if currentTerm.Score == model.Threshold {
		model.createSuggestKeys(term)
	}
	model.Unlock()
}

// For a given term, create the partially deleted lookup keys
func (model *Model) createSuggestKeys(term ColorWord) {
	edits := model.EditsMulti(term, model.Depth)
	for _, edit := range edits {
		skip := false
		for _, hit := range model.Suggest[edit.Id] {
			if term.Equals(hit) {
				// Already know about this one
				skip = true
				continue
			}
		}
		if !skip && len(edit.Colors) > 1 {
			model.Suggest[edit.Id] = append(model.Suggest[edit.Id], term)
		}
	}
}

// Edits at any depth for a given term. The depth of the model is used
func (model *Model) EditsMulti(term ColorWord, depth int) []ColorWord {
	edits := Edits1(term)
	for {
		depth--
		if depth <= 0 {
			break
		}
		for _, edit := range edits {
			edits2 := Edits1(edit)
			for _, edit2 := range edits2 {
				edits = append(edits, edit2)
			}
		}
	}
	return edits
}

// Edits1 creates a set of terms that are 1 char delete from the input term
func Edits1(word ColorWord) []ColorWord {

	splits := []Pair{}
	for i := 0; i <= len(word.Colors); i++ {
		splits = append(splits, Pair{ColorWord{Colors: word.Colors[:i]}, ColorWord{Colors: word.Colors[i:]}})
	}

	total_set := []ColorWord{}
	for _, elem := range splits {
		//deletion
		if len(elem.str2.Colors) > 0 {
			total_set = append(total_set, ColorWord{Colors: elem.str1.Colors}, ColorWord{Colors: elem.str2.Colors[1:]})
		} else {
			total_set = append(total_set, ColorWord{Colors: elem.str1.Colors})
		}

	}
	return total_set
}

func (model *Model) score(input ColorWord) int {
	if word, ok := model.Data[input.Id]; ok {
		return word.Score
	}
	return 0
}

// From a group of potentials, work out the most likely result
func best(input ColorWord, potential map[string]*Potential) ColorWord {
	best := ColorWord{}
	bestcalc := 0
	for i := 0; i < 4; i++ {
		for _, pot := range potential {
			if pot.Leven == 0 {
				return pot.Term
			} else if pot.Leven == i {
				if pot.Score > bestcalc {
					bestcalc = pot.Score
					// If the first letter is the same, that's a good sign. Bias these potentials
					if pot.Term.Colors[0] == input.Colors[0] {
						bestcalc += bestcalc * 100
					}

					best = pot.Term
				}
			}
		}
		if bestcalc > 0 {
			return best
		}
	}

	return best
}

// Test an input, if we get it wrong, look at why it is wrong. This
// function returns a bool indicating if the guess was correct as well
// as the term it is suggesting. Typically this function would be used
// for testing, not for production
func (model *Model) CheckKnown(input ColorWord, correct ColorWord) bool {
	model.RLock()
	defer model.RUnlock()
	suggestions := model.suggestPotential(input, true)
	best := best(input, suggestions)
	if best.Equals(correct) {
		// This guess is correct
		fmt.Printf("Input correctly maps to correct term")
		return true
	}
	if pot, ok := suggestions[correct.Id]; !ok {

		if model.score(correct) > 0 {
			fmt.Printf("\"%v\" - %v (%v) not in the suggestions. (%v) best option.\n", input, correct, model.score(correct), best)
			for _, sugg := range suggestions {
				fmt.Printf("	%v\n", sugg)
			}
		} else {
			fmt.Printf("\"%v\" - Not in dictionary\n", correct)
		}
	} else {
		fmt.Printf("\"%v\" - (%v) suggested, should however be (%v).\n", input, suggestions[best.Id], pot)
	}
	return false
}

// For a given input term, suggest some alternatives. If exhaustive, each of the 4
// cascading checks will be performed and all potentials will be sorted accordingly
func (model *Model) suggestPotential(input ColorWord, exhaustive bool) map[string]*Potential {
	suggestions := make(map[string]*Potential, 20)

	// 0 - If this is a dictionary term we're all good, no need to go further
	if model.score(input) > model.Threshold {
		suggestions[input.Id] = &Potential{Term: input, Score: model.score(input), Leven: 0, Method: 0}
		if !exhaustive {
			return suggestions
		}
	}

	// 1 - See if the input matches a "suggest" key
	if sugg, ok := model.Suggest[input.Id]; ok {
		for _, pot := range sugg {
			if _, ok := suggestions[pot.Id]; !ok {
				suggestions[pot.Id] = &Potential{Term: pot, Score: model.score(pot), Leven: Levenshtein(&input, &pot), Method: 1}
			}
		}

		if !exhaustive {
			return suggestions
		}
	}

	// 2 - See if edit1 matches input
	max := 0
	edits := model.EditsMulti(input, model.Depth)
	for _, edit := range edits {
		score := model.score(edit)
		if score > 0 && len(edit.Colors) > 2 {
			if _, ok := suggestions[edit.Id]; !ok {
				suggestions[edit.Id] = &Potential{Term: edit, Score: score, Leven: Levenshtein(&input, &edit), Method: 2}
			}
			if score > max {
				max = score
			}
		}
	}
	if max > 0 {
		if !exhaustive {
			return suggestions
		}
	}

	// 3 - No hits on edit1 distance, look for transposes and replaces
	// Note: these are more complex, we need to check the guesses
	// more thoroughly, e.g. levals=[valves] in a raw sense, which
	// is incorrect
	for _, edit := range edits {
		if sugg, ok := model.Suggest[edit.Id]; ok {
			// Is this a real transpose or replace?
			for _, pot := range sugg {
				lev := Levenshtein(&input, &pot)
				if lev <= model.Depth+1 { // The +1 doesn't seem to impact speed, but has greater coverage when the depth is not sufficient to make suggestions
					if _, ok := suggestions[pot.Id]; !ok {
						suggestions[pot.Id] = &Potential{Term: pot, Score: model.score(pot), Leven: lev, Method: 3}
					}
				}
			}
		}
	}
	return suggestions
}

// Return the raw potential terms so they can be ranked externally
// to this package
func (model *Model) Potentials(input ColorWord, exhaustive bool) map[string]*Potential {
	model.RLock()
	defer model.RUnlock()
	return model.suggestPotential(input, exhaustive)
}

// For a given input string, suggests potential replacements
func (model *Model) Suggestions(input ColorWord, exhaustive bool) []ColorWord {
	model.RLock()
	suggestions := model.suggestPotential(input, exhaustive)
	model.RUnlock()
	output := make([]ColorWord, 0, 10)
	for _, suggestion := range suggestions {
		output = append(output, suggestion.Term)
	}
	return output
}

// Return the most likely correction for the input term
func (model *Model) SpellCheck(input ColorWord) ColorWord {
	model.RLock()
	suggestions := model.suggestPotential(input, false)
	model.RUnlock()
	return best(input, suggestions)
}

func SampleEnglish() []string {
	var out []string
	file, err := os.Open("data/big.txt")
	if err != nil {
		fmt.Println(err)
		return out
	}
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	// Count the words.
	count := 0
	for scanner.Scan() {
		exp, _ := regexp.Compile("[a-zA-Z]+")
		words := exp.FindAll([]byte(scanner.Text()), -1)
		for _, word := range words {
			if len(word) > 1 {
				out = append(out, strings.ToLower(string(word)))
				count++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading input:", err)
	}

	return out
}

// Takes the known dictionary listing and creates a suffix array
// model for these terms. If a model already existed, it is discarded
func (model *Model) updateSuffixArr() {
	if !model.UseAutocomplete {
		return
	}
	model.RLock()
	termArr := make([]ColorWord, 0, 1000)
	for _, term := range model.Data {
		if term.Score > model.Threshold {
			termArr = append(termArr, term)
		}
	}
	//model.SuffixArrConcat = "\x00" + strings.Join(termArr, "\x00") + "\x00"
	model.SuffixArrConcat = "\x00\x00"
	model.SuffixArr = suffixarray.New([]byte(model.SuffixArrConcat))
	model.SuffDivergence = 0
	model.RUnlock()
}

// For a given string, autocomplete using the suffix array model
/*func (model *Model) Autocomplete(input ColorWord) ([]string, error) {
	fmt.Println("Autocomplete")
	if !model.UseAutocomplete {
		return []string{}, errors.New("Autocomplete is disabled")
	}
	express := "\x00" + input + "[^\x00]*"
	match, err := regexp.Compile(express)
	if err != nil {
		return []string{}, err
	}
	fmt.Println(match)
	matches := model.SuffixArr.FindAllIndex(match, -1)
	a := &Autos{Results: make([]string, 0, len(matches)), Model: model}
	for _, m := range matches {
		str := strings.Trim(model.SuffixArrConcat[m[0]:m[1]], "\x00")
		if count, ok := model.Data[str.Id]; ok && count > model.Threshold && count < model.Maxcount/50 {
			a.Results = append(a.Results, str)
		}
	}
	sort.Sort(a)
	if len(a.Results) >= 10 {
		return a.Results[:10], nil
	}
	return a.Results, nil
}*/