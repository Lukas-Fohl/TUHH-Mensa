package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const MENSA_LINK = "https://www.stwhh.de/gastronomie/mensen-cafes-weiteres/mensa/mensa-harburg"
const NTFY_LINK = "https://ntfy.sh/elene_essen"

func makeHTTPRequest(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func getTitles(stringIn string) []string {
	output := []string{}
	lineSplit := strings.Split(stringIn, "\n")
	for idx, dat := range lineSplit {
		if idx < len(lineSplit)-1 && strings.Contains(dat, "h5") && !strings.Contains(dat, "</h5>") {
			tempLine := strings.TrimSpace(lineSplit[idx+1])
			tempLine = html.UnescapeString(tempLine)
			output = append(output, tempLine)
		}
	}

	return output
}

func getPrices(stringIn string) []string {
	output := []string{}
	lineSplit := strings.Split(stringIn, "\n")
	for idx, dat := range lineSplit {
		if idx < len(lineSplit)-1 && strings.Contains(dat, "<span class=\"singlemeal__info--semibold\">") {
			tempLine := strings.ReplaceAll(lineSplit[idx+1], "&#8364;", "")
			tempLine = strings.TrimSpace(tempLine)
			_, err := strconv.ParseFloat(strings.ReplaceAll(tempLine, ",", "."), 32)
			if err == nil {
				output = append(output, tempLine)
			}
		}
	}

	return output
}

func getHTMLElement(stringIn string) []string {
	output := []string{""}
	lineSplit := strings.Split(stringIn, "\n")
	recording := false
	elementIdx := 0
	for idx, dat := range lineSplit {
		if idx == len(lineSplit)-1 {
			break
		}

		if strings.Contains(lineSplit[idx+1], "class=\"singlemeal \"") {
			recording = true
		}

		if recording {
			output[elementIdx] += dat + "\n"
		}

		if lineSplit[idx] == "</div>" {
			if recording {
				output = append(output, "")
				elementIdx++
			}
			recording = false
		}
	}

	return output
}

func removeParen(stringIn string) string {
	tempLine := ""
	depth := 0
	for _, dat := range stringIn {
		switch dat {
		case '(':
			depth++
		case ')':
			depth--
		default:
			if depth <= 0 {
				tempLine += string(dat)
			}
		}
	}

	tempLine = strings.ReplaceAll(tempLine, "  ", " ")
	tempLine = strings.ReplaceAll(tempLine, " ,", ",")
	tempLine = strings.TrimSuffix(tempLine, " ")
	return tempLine
}

func checkSchnitzel(stringIn []string, counter int) int {
	tempCount := counter
	temp := strings.Join(stringIn, ",")
	strTrans := strings.ToLower(temp)
	if strings.Contains(strTrans, "putenschnitzel") {
		tempCount++
	}
	return tempCount
}

func trackSchnitzel(stringIn []string, day int) string {
	counter := 0
	dayCheck := 0
	fileIsNew := false

	fName := "counter.schnitzel"
	content, err := os.ReadFile(fName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File '%s' doesn't exist, creating it now\n", fName)
			fileIsNew = true
		} else {
			fmt.Printf("Error reading file: %v\n", err)
			return strconv.Itoa(counter)
		}
	}

	if !fileIsNew {
		var parseErr error

		data := strings.Split(string(content), ";")
		if len(data) != 2 {
			fmt.Println("Something went wrong with the file, returning default value for counter")
			return strconv.Itoa(counter)
		}

		counter, parseErr = strconv.Atoi(data[0])
		CheckConversion(parseErr)

		dayCheck, parseErr = strconv.Atoi(data[1])
		CheckConversion(parseErr)
	}

	if dayCheck == day {
		return strconv.Itoa(counter)
	}

	augmentedCounter := checkSchnitzel(stringIn, counter)

	updatedContent := []byte(strconv.Itoa(augmentedCounter) + ";" + strconv.Itoa(day))
	err = os.WriteFile(fName, updatedContent, 0666)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
	}

	return strconv.Itoa(augmentedCounter)
}

func CheckConversion(err error) {
	if err != nil {
		fmt.Printf("Failed to convert a value: %v\n", err)
		panic(err)
	}
}

func main() {
	dat, err := makeHTTPRequest(MENSA_LINK)
	if err != nil {
		panic(err)
	}

	/*
		dat, err := os.ReadFile("out.html")
		if err != nil {
			panic(err)
		}
	*/

	lines := strings.Split(string(dat), "\n")
	finishedLines := []string{}
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			finishedLines = append(finishedLines, line)
		}
	}
	newDat := strings.Join(finishedLines, "\n")

	foodTitles := getTitles(newDat)
	if len(foodTitles) == 0 {
		return
	}

	foodPrices := getPrices(newDat)

	addTitles := false
	if len(foodTitles)*3 == len(foodPrices) {
		addTitles = true
	}

	outline := ""

	for i, dat := range foodTitles {
		outline += "• " + removeParen(dat)
		if addTitles {
			outline += " - " + foodPrices[i*3] + "€"
		}
		if i < len(foodTitles)-1 {
			outline += "\n"
		}
	}

	now := time.Now()
	day := now.Day()
	month := int(now.Month())
	title := "🍽️ TUHH-Speiseplan " + strconv.Itoa(day) + "." + strconv.Itoa(month) +
		"   Wie oft gab es schon schnitzel? : " + trackSchnitzel(foodTitles, day) + " 🤯"
	//fmt.Println(title)
	//fmt.Println(outline)

	req, _ := http.NewRequest("POST", NTFY_LINK,
		strings.NewReader(outline))
	req.Header.Set("Title", title)
	http.DefaultClient.Do(req)
}
