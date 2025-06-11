package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const MENSA_LINK = "https://www.stwhh.de/speiseplan?l=158&t=today"
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
			tempLine := strings.Trim(lineSplit[idx+1], " \t")
			output = append(output, tempLine)
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
			if recording == true {
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
			break
		case ')':
			depth--
			break
		default:
			if depth <= 0 {
				tempLine += string(dat)
			}
			break
		}
	}

	tempLine = strings.ReplaceAll(tempLine, "  ", " ")
	tempLine = strings.ReplaceAll(tempLine, " ,", ",")
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

	outline := ""
	foodTitles := getTitles(string(dat))
	if len(foodTitles) == 0 {
		return
	}

	for _, dat := range foodTitles {
		outline += "•" + removeParen(dat) + "\n"
	}

	now := time.Now()
	day := now.Day()
	month := int(now.Month())
	title := "🍽️ TUHH-Speiseplan " + strconv.Itoa(day) + "." + strconv.Itoa(month)

	req, _ := http.NewRequest("POST", NTFY_LINK,
		strings.NewReader(outline))
	req.Header.Set("Title", title)
	http.DefaultClient.Do(req)
}
