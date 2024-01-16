package kanji

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	cacheFilename, cacheWordsFilename string
)

type KanjiInfo struct {
	Kanji        string   `json:"kanji"`
	Grade        int      `json:"grade"`
	StrokeCount  int      `json:"stroke_count"`
	Meanings     []string `json:"meanings"`
	KunReadings  []string `json:"kun_readings"`
	OnReadings   []string `json:"on_readings"`
	NameReadings []string `json:"name_readings"`
	JLPT         int      `json:"jlpt"`
	Unicode      string   `json:"unicode"`
	EN           string   `json:"heisig_en"`
}

func init() {
	var err error

	// UserConfigDir returns the default root directory to use for user-specific configuration data. Users should create their own application-specific subdirectory within this one and use that.
	// On Unix systems, it returns $XDG_CONFIG_HOME as specified by https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html if non-empty, else $HOME/.config. On Darwin, it returns $HOME/Library/Application Support. On Windows, it returns %AppData%. On Plan 9, it returns $home/lib.
	// If the location cannot be determined (for example, $HOME is not defined), then it will return an error.
	// C:\Users\YourUser\AppData\Roaming
	cnfpath, err := os.UserConfigDir()
	check(err)

	// Setup log and cache files
	cacheFilename = fixPath(fmt.Sprintf(`%v/kanji_lookup/kanji_cache.json`, cnfpath))
	cacheWordsFilename = fixPath(fmt.Sprintf(`%v/kanji_lookup/word_cache.json`, cnfpath))
	// logFilename := fixPath(fmt.Sprintf(`%v/kanji_lookup.log`, cnfpath))
	log.Printf("kanji cache: %v", cacheFilename)
	// log.Printf("kanji log: %v", logFilename)
}

func Lookup(text string) string {

	// start script
	log.Printf("Kanji Lookup: %v\n", text)
	// cache := loadCache(cacheFilename)
	// keys := make([]string, 0, len(cache))
	// for k := range cache {
	// 	keys = append(keys, k)
	// }
	// fmt.Println(keys)
	// return ""
	// extract all unique kanji
	unique := uniqueSlice(text)
	unique = strings.ReplaceAll(unique, "\n", "")

	// log.Println("Getting Kanji Info")
	result := getKanjiInfo(cacheFilename, unique)

	return result
}

func lookupKanji(kanji string) KanjiInfo {

	endpoint := fmt.Sprintf("https://kanjiapi.dev/v1/kanji/%v", kanji)
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatalln(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	check(err)
	// body := string(data)

	if resp.StatusCode != 200 {
		log.Printf("Querying: %v", endpoint)
		log.Println(resp.StatusCode)
		log.Println(string(data))
	}

	var info KanjiInfo
	err = json.Unmarshal(data, &info)
	check(err)

	return info
}

func loadCache(filename string) map[string]KanjiInfo {

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("no cache found: %v\n", filename)
		return make(map[string]KanjiInfo)
	}

	// log.Println("Loading Cache")
	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	check(err)

	if !json.Valid(data) {
		log.Printf("cache file corrupted: %v\n", filename)
		return make(map[string]KanjiInfo)
	}

	var list map[string]KanjiInfo
	err = json.Unmarshal(data, &list)
	check(err)

	return list
}

func infoToString(info KanjiInfo) string {
	return fmt.Sprintf("%v - on(%v) - kun(%v): %v",
		info.Kanji,
		strings.Join(info.OnReadings, ", "),
		strings.Join(info.KunReadings, ", "),
		strings.Join(info.Meanings, "; "))
}

func getKanjiInfo(cacheFilename string, kanjiList string) string {
	var err error
	datawriter := bytes.NewBufferString("")

	cache := loadCache(cacheFilename)
	for _, value := range kanjiList {
		key := string(value)

		var info KanjiInfo
		// if the kanji doesn't exist in the cache, get it
		if _, ok := cache[key]; !ok {
			info = lookupKanji(key)
			cache[key] = info
		} else {
			info = cache[key]
		}

		_, err = datawriter.WriteString(infoToString(info) + "\n")
		check(err)
	}

	saveCache(cacheFilename, cache)

	return datawriter.String()
}

func saveCache(filename string, cache map[string]KanjiInfo) {

	// log.Println("Saving Cache")
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)

	cacheData, err := json.Marshal(cache)
	check(err)

	cacheWriter := bufio.NewWriter(file)
	_, err = cacheWriter.Write(cacheData)
	check(err)
}

/////////////////////////////////////////////////////////////

func uniqueSlice(s string) string {
	text := sortString(s)
	keys := make(map[rune]bool)
	list := ""
	for _, entry := range text {
		if isNotKanji(entry) {
			continue
		}
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = list + string(entry)
		}
	}
	return list
}
