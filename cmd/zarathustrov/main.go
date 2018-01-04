package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

var endings = map[rune]bool{
	'.': true,
	'!': true,
	'?': true,
}

var seperators = map[rune]bool{
	',': true,
	';': true,
}

// randInt returns a random integer between a max and a min
// see https://stackoverflow.com/q/12321133
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// parseCorpus takes a File of text, and creates the Markov Tree. It builds the
// tree by parsing each rune, determining if it is an English letter, and
// assigning it to a nested map. It also tracks the frequency of appearances
// of punctuation following the word, as well as the total number of
// appearances of that word in the corpus. The function returns a pointer to
// the map. It also requires the starting and ending character of the file
// to be read. If '-1' is supplied to the last character, all characters are
// read.
// Note that the tree in JSON form as tree.json in the working directory.
func parseCorpus(corpus *os.File, start int, end int) *map[string]map[string]int {
	// read the entire text file as an array.
	// TODO: chunk it, use multiple gothreads?
	text, err := ioutil.ReadAll(corpus)
	if err != nil {
		log.Fatal(err)
	}
	if end == -1 {
		end = len(text)
	}
	content := string(text[start:end])

	tree := make(map[string]map[string]int)
	word := []rune("the") // TODO: Get real first word
	nextWord := []rune("")

	for _, letter := range content {
		// check if it's really a letter or something else
		// if it's punctuation, it will get added later
		if (letter >= rune('a') && letter <= rune('z')) ||
			(letter >= rune('A') && letter <= rune('Z')) {
			nextWord = append(nextWord, letter)
		} else {
			key := strings.ToLower(string(word))
			value := strings.ToLower(string(nextWord))
			// check if one of the words is blank. If so, skip
			if len(key) == 0 || len(value) == 0 {
				continue
			}
			// Reset the words
			word = nextWord
			nextWord = make([]rune, 0)

			if _, ok := tree[key]; !ok {
				tree[key] = make(map[string]int)
				// Keep track of the overall count per entry,
				// to make it easy to probabilities
				tree[key][value] = 1
				tree[key]["_appearances"] = 1
			} else {
				count := tree[key][value]
				tree[key][value] = count + 1
				count = tree[key]["_appearances"]
				tree[key]["_appearances"] = count + 1
			}

			if endings[letter] || seperators[letter] {
				punctuation := string(letter)
				if _, ok := tree[value]; !ok {
					tree[value] = make(map[string]int)
					tree[value][punctuation] = 1
					tree[value]["_appearances"] = 1
				} else {
					count := tree[value][punctuation]
					tree[value][punctuation] = count + 1
					count = tree[value]["_appearances"]
					tree[value]["_appearances"] = count + 1
				}
			}
		}
	}

	jsonString, err := json.Marshal(tree)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("assets/tree.json", jsonString, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return &tree
}

// generateRandomString takes a Markov tree map of string[int] maps, and
// returns a single string less than 280 characters (for Twitter).
func generateRandomString(tree map[string]map[string]int) string {
	var key string
	var randomString string
	var word string
	var position int
	// var quoteOpen bool TODO: Fix so that all quotes are closed
	for key = range tree {
		if word != "_appearances" {
			break
		}
	}

	randomString = strings.Title(key)
	for len(randomString) < 280 {
		secondMap := tree[key]
		nth := randInt(0, secondMap["_appearances"])
		// Ranging through maps is basically random, per
		// https://blog.golang.org/go-maps-in-action
		// (ctrl-F "Iteration order")
		// This section cycles through until nth becomes 0 or less.
		// It basically draws a random word, and whose count is then
		// subtracted from nth. This preserves the probability, since
		// nth is a random number from 0 to the total number of
		// possible second words, and each is weighted by its count
		for word = range secondMap {
			if word == "_appearances" {
				continue
			}
			nth = nth - secondMap[word]
			if nth <= 0 {
				break
			}
		}

		switch word {
		case ".", "!", "?":
			position = len(randomString) + 1
			for key = range tree {
				break
			}
			randomString = randomString + word + " " + strings.Title(key)
		case ",", ";":
			for key = range tree {
				break
			}
			randomString = randomString + word + " " + key
		case "i", "zarathustra":
			randomString = randomString + " " + strings.Title(word)
			key = word
		default:
			randomString = randomString + " " + word
			key = word
		}
	}

	// FIXME: This only runs once, what if the string never has an ending?
	// TODO: Check for open quotes
	return randomString[:position]
}

// twitterLogin returns the Twitter connected API Client
func twitterLogin() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(twitterConsumerKey)
	anaconda.SetConsumerSecret(twitterConsumerSecret)
	api := anaconda.NewTwitterApi(twitterAccessToken, twitterAccessSecret)
	return api
}

// makePost takes a Markov tree, then gets a Twitter API Client and a new
// random string and posts it
func makePost(tree *map[string]map[string]int) {
	twitterClient := twitterLogin()
	text := generateRandomString(*tree)
	twitterClient.PostTweet(text, nil)
}

// main runs the program forever, calling makePost every 15 minutes
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	corpus, err := os.Open("assets/corpus.txt")
	if err != nil {
		log.Fatal(err)
	}
	// 31271 and 545805 are magic numbers that cover the characters in
	// corpus.txt that exclude the introduction and the afterword
	tree := parseCorpus(corpus, 31271, 545805)
	corpus.Close()

	for {
		makePost(tree)
		time.Sleep(15 * time.Minute)
	}
}
