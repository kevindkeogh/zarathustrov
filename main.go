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
	text, err := ioutil.ReadAll(corpus)
	if err != nil {
		log.Fatal(err)
	}
	if end == -1 {
		end = len(text)
	}
	content := string(text[start:end])

	lookup := make(map[string]map[string]int)
	word := []rune("the") // TODO: Get real first word
	nextWord := []rune("")

	for _, letterRune := range content {
		// check if letter
		if (letterRune >= rune('a') && letterRune <= rune('z')) ||
			(letterRune >= rune('A') && letterRune <= rune('Z')) {
			nextWord = append(nextWord, letterRune)
		} else {
			key := strings.ToLower(string(word))
			value := strings.ToLower(string(nextWord))
			// check if one of the words is blank. If so, skip
			if len(key) == 0 || len(value) == 0 {
				continue
			}

			if _, ok := lookup[key]; !ok {
				lookup[key] = make(map[string]int)
			}
			count := lookup[key][value]
			lookup[key][value] = count + 1
			// Keep track of the overall count per entry,
			// to make it easy to probabilities
			count = lookup[key]["_appearances"]
			lookup[key]["_appearances"] = count + 1

			switch letterRune {
			case '.', ';', ',', '!', '?', '"':
				if _, ok := lookup[value]; !ok {
					lookup[value] = make(map[string]int)
				}
				count := lookup[value][string(letterRune)]
				lookup[value][string(letterRune)] = count + 1
				count = lookup[value]["_appearances"]
				lookup[value]["_appearances"] = count + 1
			}

			word = nextWord
			nextWord = make([]rune, 0)
		}
	}

	jsonString, _ := json.Marshal(lookup)
	err = ioutil.WriteFile("tree.json", jsonString, 0666)
	return &lookup
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
		case "i":
			randomString = randomString + " " + strings.Title(word)
			key = word
		default:
			randomString = randomString + " " + word
			key = word
		}
	}

	if position == 0 {
		randomString = generateRandomString(tree)
	}
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

func makePost(tree *map[string]map[string]int) {
	twitterClient := twitterLogin()
	text := generateRandomString(*tree)
	twitterClient.PostTweet(text, nil)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	corpus, err := os.Open("corpus.txt")
	if err != nil {
		log.Fatal(err)
	}
	tree := parseCorpus(corpus, 31271, 545805)
	corpus.Close()

	for {
		makePost(tree)
		time.Sleep(15 * time.Minute)
	}
}
