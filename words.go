package kanji

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type WordInfo struct {
	Kanji   string `json:"kanji"`
	Kana    string `json:"kana"`
	Type    string `json:"type"`
	Meaning string `json:"meaning"`
	Count   int    `json:"count"`
}

func LookupWords(text, dir, name string) map[string]WordInfo {

	scanned := ScanForWords(text)
	// id := time.Now().Format("150405")
	// saveFile(fmt.Sprintf("%v/%v_scan.html", dir, id), scanned.Bytes())

	words, _ := ScrapeHTML(scanned)
	// saveFile(fmt.Sprintf("%v/%v_words.txt", dir, name), words.Bytes())

	UpdateCache(cacheWordsFilename, words)
	return words
}

func UpdateCache(filename string, words map[string]WordInfo) map[string]WordInfo {
	cache := loadWordCache(filename)
	for key, value := range words {
		if cw, ok := cache[key]; ok {
			// fmt.Println("add existing count to new count")
			value.Count = value.Count + cw.Count
			v := cache[key]
			v.Count += 1
			cache[key] = v
			continue
		}

		cache[key] = value
	}

	saveWordCache(filename, cache)
	return cache
}

func LookupKanjiToKana(input string) string {

	var text string
	characters := []rune(input)

	// reject any english or special characters
	for _, v := range characters {
		// log.Println(string(v), v, int(v))
		if int(v) < 256 {
			continue
		}
		text += string(v)
	}

	htmlResponse := lookupKana(text)
	kanaText := ScrapeKanaConvertHTML(bytes.NewBufferString(htmlResponse))

	return kanaText
}

func ScanForWords(d string) *bytes.Buffer {

	// query for word recognition
	res := lookupWords(d)
	final := bytes.NewBufferString("")
	final.WriteString(res)

	return final
}

func ScrapeHTML(buf io.Reader) (map[string]WordInfo, *bytes.Buffer) {

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		log.Fatal(err)
	}

	unique := make(map[string]WordInfo, 0)
	contentRegex := regexp.MustCompile("^(.*)<br>([a-zA-Z]+)<br>(.*)$")

	// Find the review items
	final := bytes.NewBufferString("")
	doc.Find(".tool__results").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			// For each item found, get the title
			title := a.Text()

			// reject any dictation of particles
			characters := []rune(title)
			if len(characters) == 1 && isNotKanji(characters[0]) {
				// log.Println("DISCARDED:", title)
				return
			}

			// reject any english or special characters
			for _, v := range characters {
				// log.Println(string(v), v, int(v))
				if int(v) < 256 {
					return
				}
			}

			// determine if word has kanji or not
			// for _, v := range []rune(title) {
			// 	if isNotKanji(v) {
			// 		good = false
			// 		break
			// 	}
			// }

			content, exists := a.Attr("data-tooltip")
			if strings.Contains(content, "postposition") || strings.Contains(content, "symbol") {
				return
			}

			m, _ := regexp.Match("^\\s+$", []byte(content))
			m2, _ := regexp.Match("^\\s+$", []byte(title))
			if !exists || content == "" || m || m2 {
				return
			}
			if isNotKanji([]rune(title)[0]) {
				return
			}
			newWord := WordInfo{
				Kanji:   title,
				Meaning: strings.Replace(content, title, "", 1),
			}
			if contentRegex.MatchString(content) {
				match := contentRegex.FindStringSubmatch(content)
				// fmt.Println("matches:", match)
				if len(match) >= 4 {
					newWord.Kana = match[1]
					newWord.Type = match[2]
					newWord.Meaning = match[3]
					newWord.Count = 1
					// fmt.Println("found match")
				}
			} else {
				fmt.Println("---no matched content---")
				fmt.Println(content)
				fmt.Println("------------------")
			}
			content = strings.ReplaceAll(content, "<br>", " | ")
			unique[title] = newWord

			final.WriteString(fmt.Sprintf("%v:%v\n", title, content))
		})
	})

	return unique, final
}

func ScrapeKanaConvertHTML(buf io.Reader) string {

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		log.Fatal(err)
	}

	var outputText string

	// Find the review items
	doc.Find(".tool__results").Each(func(i int, s *goquery.Selection) {
		outputText += strings.TrimSpace(s.Text())
	})

	return outputText
}

func saveWordCache(filename string, cache map[string]WordInfo) {

	// log.Println("Saving Cache")
	cacheData, err := json.Marshal(cache)
	check(err)

	// fmt.Println("CACHED DATA")
	// fmt.Println(string(cacheData))

	saveFile(filename, cacheData)
}

func loadWordCache(filename string) map[string]WordInfo {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("no cache found: %v\n", filename)
		return make(map[string]WordInfo)
	}

	// log.Println("Loading Cache")
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	check(err)

	if !json.Valid(data) {
		log.Printf("cache file corrupted: %v\n", filename)
		return make(map[string]WordInfo)
	}

	var list map[string]WordInfo
	err = json.Unmarshal(data, &list)
	check(err)

	return list
}
