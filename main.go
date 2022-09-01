package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	// ScanFiles([]string{".\\scripts\\"})
	// ScrapeFile("./results/as_re_1_script.txt.html")
	ScrapeFiles("./results/")
}

/**
 * arguments:
 * 1: path to script file
 */
func Run(args []string) {

	// verify path was passed to the script
	if len(args) < 1 {
		panic(fmt.Errorf("kanji lookup failed, required arguments: <path_to_script>"))
	}

	dir := args[0]
	outputFile := fixPath("./kanji.txt")
	outputFile2 := fixPath("./kanji_list.txt")

	// Verify required files
	log.Printf("Starting Kanji Lookup Script: %v\n", dir)
	log.Printf("Output: %v\n", outputFile)

	text := string(readFiles(dir))

	// count each kanji and put it into a map
	counts := countKanji(text)
	keys := make([]rune, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}

	// sort the list
	sort.SliceStable(keys, func(i, j int) bool {
		return counts[keys[i]] > counts[keys[j]]
	})

	i := 0
	final := bytes.NewBufferString("")
	finalList := bytes.NewBufferString("")
	for _, k := range keys {
		i++

		kanji := string(rune(k))

		// write text to buffer
		final.WriteString(fmt.Sprintf("%v: %v\n", kanji, counts[k]))
		finalList.WriteString(kanji)
		if i%20 == 0 {
			finalList.WriteString("\n")
		}
	}

	// remove any old output files to prevent corrupted results
	if _, err := os.Stat(outputFile); os.IsExist(err) {
		e := os.Remove(outputFile)
		if e != nil {
			log.Fatal(e)
		}
	}

	// create output file
	out, err := os.Create(outputFile)
	check(err)
	_, err = out.Write(final.Bytes())
	check(err)
	out.Close()

	// save list
	out, err = os.Create(outputFile2)
	check(err)
	_, err = out.Write(finalList.Bytes())
	check(err)
	out.Close()

	log.Println("Kanji Count Success!")
}

func ScrapeFiles(dir string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	unique := make(map[string]string, 0)
	for _, f := range files {
		inputFile := fmt.Sprintf("%v/%v", dir, f.Name())
		// outputFile := fmt.Sprintf("words/%v.txt", f.Name())
		log.Println("file: ", inputFile)

		u := ScrapeFile(inputFile)
		for k, v := range u {
			// its fine to overwrite since the results will be identical
			unique[k] = v
		}
	}

	// save uniques
	final := bytes.NewBufferString("")
	for k, v := range unique {
		final.WriteString(fmt.Sprintf("%v:%v\n", k, v))
	}

	out, err := os.Create("unique_words.txt")
	check(err)
	_, err = out.Write(final.Bytes())
	check(err)
	out.Close()
}
func ScrapeFile(filename string) map[string]string {
	name := filepath.Base(filename)
	outputFile := fmt.Sprintf("./words/%v.txt", name)

	// Request the HTML page.
	res, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewBuffer(res)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		log.Fatal(err)
	}

	unique := make(map[string]string, 0)

	// Find the review items
	final := bytes.NewBufferString("")
	doc.Find(".tool__results").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			// For each item found, get the title
			title := a.Text()
			good := false

			// reject any english words
			for _, v := range []rune(title) {
				if !isNotKana(v) || !isNotKanji(v) {
					good = true
					break
				}
			}
			if !good {
				return
			}

			// determine if word has kanji or not
			for _, v := range []rune(title) {
				if isNotKanji(v) {
					good = false
					break
				}
			}

			content, exists := a.Attr("data-tooltip")
			m, _ := regexp.Match("^\\s+$", []byte(content))
			m2, _ := regexp.Match("^\\s+$", []byte(title))
			if !exists || content == "" || m || m2 {
				return
			}
			content = strings.ReplaceAll(content, "<br>", " | ")
			if good {
				unique[title] = strings.Replace(content, title, "", 1)
			}

			final.WriteString(fmt.Sprintf("%v\n", content))
		})
	})

	// remove any old output files to prevent corrupted results
	if _, err := os.Stat(outputFile); os.IsExist(err) {
		e := os.Remove(outputFile)
		if e != nil {
			log.Fatal(e)
		}
	}

	out, err := os.Create(outputFile)
	check(err)
	_, err = out.Write(final.Bytes())
	check(err)
	out.Close()

	return unique
}

/**
 * arguments:
 * 1: path to script file
 */
func ScanFiles(args []string) {

	// verify path was passed to the script
	if len(args) < 1 {
		panic(fmt.Errorf("kanji lookup failed, required arguments: <path_to_script>"))
	}

	dir := args[0]

	log.Println("reading dir: ", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	final := bytes.NewBufferString("")
	for _, f := range files {
		p := fmt.Sprintf("%v/%v", dir, f.Name())
		outputFile := fmt.Sprintf("results/%v.html", f.Name())
		log.Println("file: ", p)

		d, err := ioutil.ReadFile(p)
		check(err)

		res := lookup(string(d))
		final.WriteString(res)
		time.Sleep(time.Second * 2)

		// remove any old output files to prevent corrupted results
		if _, err := os.Stat(outputFile); os.IsExist(err) {
			e := os.Remove(outputFile)
			if e != nil {
				log.Fatal(e)
			}
		}

		// create output file
		out, err := os.Create(outputFile)
		check(err)
		_, err = out.Write(final.Bytes())
		check(err)
		out.Close()
	}
}

func ScanFile(filename string) {

	dir := filepath.Dir(filename)
	name := filepath.Base(filename)
	outputFile := fmt.Sprintf("results/%v.html", name)
	final := bytes.NewBufferString("")

	log.Println("File: ", filename)
	log.Println("Directory: ", dir)

	d, err := ioutil.ReadFile(filename)
	check(err)

	res := lookup(string(d))
	final.WriteString(res)

	// remove any old output files to prevent corrupted results
	if _, err := os.Stat(outputFile); os.IsExist(err) {
		e := os.Remove(outputFile)
		if e != nil {
			log.Fatal(e)
		}
	}

	// create output file
	out, err := os.Create(outputFile)
	check(err)
	_, err = out.Write(final.Bytes())
	check(err)
	out.Close()
}

func sortString(w string) string {
	s := strings.Split(w, "")
	sort.Strings(s)
	return strings.Join(s, "")
}

func isNotKana(k rune) bool {
	min := 0x3040
	max := 0x309f

	return int(k) < min || int(k) > max
}

func isNotKanji(k rune) bool {
	min := 0x4e00
	max := 0x9faf

	return int(k) < min || int(k) > max
}

func countKanji(s string) map[rune]int {
	text := sortString(s)
	keys := make(map[rune]int)
	for _, entry := range text {
		if isNotKanji(entry) {
			continue
		}
		if c, value := keys[entry]; value {
			keys[entry] = c + 1
		} else {
			keys[entry] = 1
		}
	}
	return keys
}

func checkFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Errorf("file does not exist, aborting script: %v", path))
	}
}

func readFiles(dir string) []byte {
	var dat []byte
	var err error

	log.Println("reading dir: ", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		p := fmt.Sprintf("%v/%v", dir, f.Name())
		log.Println("file: ", p)

		d, err := ioutil.ReadFile(p)
		check(err)

		dat = append(dat, d...)
	}

	return dat
}

func fixPath(p string) string {
	return filepath.ToSlash(path.Clean(p))
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

/*
curl -X POST https://nihongodera.com/tools/convert
   -H "Content-Type: application/x-www-form-urlencoded"
   -d "options[analyzer][]=analyzer&options[analyzer][words]=words&_token=YyzE1A3yvFE2g7yBpY8oNN7Z5yaO0rtZTZUf4Pqc&text=たしかに場所は伝えましたが&type=analyzer"
*/
func lookup(kanji string) string {

	endpoint := "https://nihongodera.com/tools/convert"
	q := fmt.Sprintf("options[analyzer][]=analyzer&options[analyzer][words]=words&_token=YyzE1A3yvFE2g7yBpY8oNN7Z5yaO0rtZTZUf4Pqc&text=%v&type=analyzer", kanji)
	buf := bytes.NewBufferString(q)
	log.Printf("Querying: %v", endpoint)
	resp, err := http.Post(endpoint, "application/x-www-form-urlencoded", buf)
	if err != nil {
		log.Fatalln(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	check(err)
	body := string(data)

	log.Println(resp.StatusCode)
	// log.Println(body)

	return body
}
