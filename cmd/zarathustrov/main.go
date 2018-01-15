package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/ChimeraCoder/anaconda"
)

type Node map[string]int
type Tree map[string]*Node

var endings = map[rune]bool{
	'.': true,
	'!': true,
	'?': true,
}

var seperators = map[rune]bool{
	',': true,
	';': true,
	':': true,
}

// randInt returns a random integer between a max and a min
// see https://stackoverflow.com/q/12321133
func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// newTree returns a new Tree struct
func newTree() *Tree {
	tree := make(Tree)
	tree["_appearances"] = newNode()
	return &tree
}

// newTree returns an empty Node
func newNode() *Node {
	node := make(Node)
	return &node
}

// update method for Tree increments the total number of words for the tree,
// then increments the key-value pair supplied. The key and value are added
// to the Tree and Node, respectively, if they are new to the Tree.
func (tree *Tree) update(key string, value string) {
	// Update overall count
	count := (*(*tree)["_appearances"])["total"]
	(*(*tree)["_appearances"])["total"] = count + 1

	// check if the key is in the tree and add to the appearances
	if _, ok := (*tree)[key]; ok {
		count := (*(*tree)[key])["_appearances"]
		(*(*tree)[key])["_appearances"] = count + 1
	} else {
		(*tree)[key] = newNode()
		(*(*tree)[key])["_appearances"] = 1
	}

	count = (*(*tree)[key])[value]
	(*(*tree)[key])[value] = count + 1
}

// getRandomKey method for a tree returns a random key from the tree.
// Note that the keys are weighted by the total number of words in the corpus.
// Ranging through maps is basically random, per
// https://blog.golang.org/go-maps-in-action (ctrl-F "Iteration order")
// This function cycles through until nth becomes 0 or less. It basically
// draws a random word, and whose count is then subtracted from nth. This
// preserves the probability, since nth is a random number from 0 to the total
// number of possible second words, and each is weighted by its count.
func (tree *Tree) getRandomKey() string {
	var key string
	nth := randInt(0, (*(*tree)["_appearances"])["total"])
	for key = range *tree {
		if key == "_appearances" {
			continue
		}
		nth = nth - (*(*tree)[key])["_appearances"]
		if nth <= 0 {
			break
		}
	}
	return key
}

// getRandomKey method for a node and returns a random key from a node.
// Note that the keys are weighted by the total number of words appearing after
// that Node in the tree.
// if omitPunct is true, the function will not return an ending or a seperator
// and will adjust the weighting to ignore punctuation
func (node *Node) getRandomKey(omitPunct bool) string {
	var key string
	var letter rune

	max := (*node)["_appearances"]
	if omitPunct {
		for k := range endings {
			max = max - (*node)[string(k)]
		}
		for k := range seperators {
			max = max - (*node)[string(k)]
		}
	}

	nth := randInt(0, max)
	for key = range *node {
		letter = []rune(key)[0]
		if key == "_appearances" ||
			(omitPunct && (endings[letter] || seperators[letter])) {
			continue
		}

		nth = nth - (*node)[key]
		if nth <= 0 {
			break
		}
	}
	return key
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
func parseCorpus(corpus *os.File, start int, end int) *Tree {
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

	tree := newTree()

	var setWord bool
	var word []rune
	var nextWord []rune

	for _, letter := range content {
		// check if it's really a letter or something else
		// if it's punctuation, it will get added later
		if (letter >= rune('a') && letter <= rune('z')) ||
			(letter >= rune('A') && letter <= rune('Z')) {
			nextWord = append(nextWord, letter)
		} else {
			if len(word) == 0 || len(nextWord) == 0 {
				if unicode.IsSpace(letter) && !setWord {
					setWord = true
					word = nextWord
					nextWord = make([]rune, 0)
				}
				continue
			}

			key := strings.ToLower(string(word))
			value := strings.ToLower(string(nextWord))

			// Update the tree
			tree.update(key, value)
			if endings[letter] || seperators[letter] {
				tree.update(value, string(letter))
			}

			// Reset the words
			word = nextWord
			nextWord = make([]rune, 0)
		}
	}

	jsonString, err := json.MarshalIndent(tree, "", "    ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("assets/tree.json", jsonString, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return tree
}

// generateRandomString takes a Markov tree map of string[int] maps, and
// returns a single string less than 280 characters (for Twitter).
func (tree *Tree) generateRandomString() string {
	var position int

	key := tree.getRandomKey()
	randomString := strings.Title(key)
	for len(randomString) < 280 {
		word := (*tree)[key].getRandomKey(false)
		letter := []rune(word)[0]
		switch {
		case endings[letter]:
			position = len(randomString) + 1
			key = tree.getRandomKey()
			randomString = randomString + word + " " + strings.Title(key)
		case seperators[letter]:
			key = (*tree)[key].getRandomKey(true)
			randomString = randomString + word + " " + key
		case word == "i" || word == "zarathustra":
			randomString = randomString + " " + strings.Title(word)
			key = word
		default:
			randomString = randomString + " " + word
			key = word
		}
	}
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
func makePost(tree *Tree) {
	var text string
	twitterClient := twitterLogin()
	for text == "" {
		text = tree.generateRandomString()
	}

	_, err := twitterClient.PostTweet(text, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// main runs the program forever, calling makePost every 15 minutes
func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	corpus, err := os.Open("assets/corpus.txt")
	if err != nil {
		log.Fatal(err)
	}
	// 32328 and 545805 are magic numbers that cover the characters in
	// corpus.txt that exclude the introduction and the afterword
	tree := parseCorpus(corpus, 32328, 564346)
	corpus.Close()

	for {
		makePost(tree)
		time.Sleep(60 * time.Minute)
	}
}
