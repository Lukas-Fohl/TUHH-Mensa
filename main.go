package main

import (
	"fmt"
	"html"
	"io"
	"net/http"
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
	title := "🍽️ TUHH-Speiseplan " + strconv.Itoa(day) + "." + strconv.Itoa(month)
	//fmt.Println(title)
	//fmt.Println(outline)

	req, _ := http.NewRequest("POST", NTFY_LINK,
		strings.NewReader(outline))
	req.Header.Set("Title", title)
	http.DefaultClient.Do(req)
}
