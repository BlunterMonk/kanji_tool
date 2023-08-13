package kanji

import (
	"bytes"
	"fmt"
	"log"
	"sort"
)

/**
 * arguments:
 * 1: path to script file
 */
func Count(args []string) {

	// verify path was passed to the script
	if len(args) < 1 {
		panic(fmt.Errorf("kanji lookup failed, required arguments: <path_to_script>"))
	}

	scriptFilename := fixPath(args[0])
	outputFile := "./counts.txt"
	text := string(readFiles(scriptFilename))

	log.Printf("Kanji Count Directory: %v\n", scriptFilename)
	// log.Printf("Kanji Count: %v\n", text)

	// unique := uniqueSlice(text)
	// unique = strings.ReplaceAll(unique, "\n", "")

	counts := countKanji(text)

	keys := make([]rune, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return counts[keys[i]] > counts[keys[j]]
	})

	values := bytes.NewBufferString("")
	final := bytes.NewBufferString("")
	uncommon := bytes.NewBufferString("")
	common := bytes.NewBufferString("")
	for i := 0; i < len(counts); i++ {
		key := keys[i]

		values.WriteString(fmt.Sprintf("%v: %v\n", string(key), counts[key]))

		// write text to buffer
		if counts[key] <= 10 {
			uncommon.WriteString(string(key))
		} else if counts[key] > 10 && counts[key] < 50 {
			common.WriteString(string(key))
		} else {
			final.WriteString(string(key))
		}
	}

	saveFile(outputFile, final.Bytes())
	saveFile("./uncommon.txt", uncommon.Bytes())
	saveFile("./common.txt", common.Bytes())
	saveFile("./values.txt", values.Bytes())

	log.Println("Kanji Lookup Success!")
}
