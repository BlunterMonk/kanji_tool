package kanji

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

const (
	TOKEN = "CnDxyYHZQayHBZSbLd9apDWX8Iez0c0XUIBqIsSG"
)

var (
	httpClient *http.Client
)

func init() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient = &http.Client{Transport: tr}
	_, err := httpClient.Get("https://nihongodera.com")
	if err != nil {
		fmt.Println(err)
	}
}

func buildPaths(filename string) (dir, name string) {
	filename = fixPath(filename)
	name = filepath.Base(filename)
	name = name[:strings.LastIndex(name, ".")]
	dir = filepath.Dir(filename)
	return dir, name
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
	min := int64(0x4e00)
	max := int64(0x9faf)
	// log.Println("KANJI CHECK:", min, "<", k, string(k), "<", max)

	return int64(k) < min || int64(k) > max
}

func CountKanji(s string) map[rune]int {
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

func readFile(file string) []byte {

	log.Println(file)
	d, err := ioutil.ReadFile(file)
	check(err)

	return d
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
		panic(e)
	}
}

/*
curl -X POST https://nihongodera.com/tools/convert
   -H "Content-Type: application/x-www-form-urlencoded"
   -d "options[analyzer][]=analyzer&options[analyzer][words]=words&_token=YyzE1A3yvFE2g7yBpY8oNN7Z5yaO0rtZTZUf4Pqc&text=たしかに場所は伝えましたが&type=analyzer"
*/
func lookupWords(text string) string {

	endpoint := "https://nihongodera.com/tools/convert"
	q := fmt.Sprintf("options[analyzer][]=analyzer&options[analyzer][words]=words&_token=%s&text=%v&type=analyzer", TOKEN, text)
	buf := bytes.NewBufferString(q)
	log.Printf("Querying: %v", endpoint)
	resp, err := httpClient.Post(endpoint, "application/x-www-form-urlencoded", buf)
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

func lookupKana(text string) string {

	endpoint := "https://nihongodera.com/tools/convert"
	q := fmt.Sprintf("options[kana][style]=hiragana&options[kana][space][type]=space&_token=%s&text=%v&type=kana", TOKEN, text)
	buf := bytes.NewBufferString(q)
	log.Printf("Querying Kana: %v", endpoint)
	resp, err := httpClient.Post(endpoint, "application/x-www-form-urlencoded", buf)
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

func saveFile(outputFile string, data []byte) {
	// remove any old output files to prevent corrupted results
	if _, err := os.Stat(outputFile); os.IsExist(err) {
		e := os.Remove(outputFile)
		if e != nil {
			log.Fatal(e)
		}
	}

	out, err := os.Create(outputFile)
	check(err)
	_, err = out.Write(data)
	check(err)
	out.Close()
}
