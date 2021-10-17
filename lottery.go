package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	hs "github.com/patrickbucher/htmlsqueeze"
	"golang.org/x/net/html"
)

const (
	lotteryURL = "https://www.irishlottery.com/daily-million-archive"
	userAgent  = "Mozilla/5.0 (X11; Linux x86_64; rv:93.0) Gecko/20100101 Firefox/93.0"

	dateInputFmt  = "2006 January 2 3:04 pm"
	dateOutputFmt = "02.01.2006 15:04"
)

var (
	datePat = regexp.MustCompile(`^([A-Z][a-z]+) ([0-9]{1,2})[a-z]{2} ([0-9]{4})`)
	timePat = regexp.MustCompile(`([0-9]{1,2}):([0-9]{2})([ap]m)`)
)

type Draft struct {
	Date      *time.Time
	Balls     []int
	ExtraBall int
}

func (d Draft) Format() string {
	buf := bytes.NewBufferString("")
	datetime := d.Date.Format(dateOutputFmt)
	buf.WriteString(datetime)
	buf.WriteRune(' ')
	for _, ball := range d.Balls {
		buf.WriteString(fmt.Sprintf("%2d ", ball))
	}
	buf.WriteString(fmt.Sprintf("Zz: %2d", d.ExtraBall))
	return buf.String()
}

func main() {
	doc, err := getDocument(lotteryURL, userAgent)
	if err != nil {
		log.Fatal(err)
	}
	for _, draft := range extractDrafts(doc) {
		fmt.Println(draft.Format())
	}
}

func extractDrafts(doc *html.Node) []Draft {
	tableRowMatcher := [][]hs.Predicate{
		[]hs.Predicate{hs.TagMatcher("tr")},
	}
	datetimeMatcher := [][]hs.Predicate{
		[]hs.Predicate{hs.TagMatcher("th")},
		[]hs.Predicate{hs.TagMatcher("a")},
	}
	ballsMatcher := [][]hs.Predicate{
		[]hs.Predicate{hs.TagMatcher("td")},
		[]hs.Predicate{hs.TagMatcher("ul")},
		[]hs.Predicate{hs.TagMatcher("li")},
	}
	nodes := hs.Apply(doc, tableRowMatcher)
	drafts := make([]Draft, 0)
	for _, node := range nodes {
		datetime := hs.Squeeze(node, datetimeMatcher, hs.ExtractChildText)
		if len(datetime) < 1 {
			continue
		}
		raw := strings.TrimSpace(datetime[0])
		drawDate := parseDate(raw)
		balls := toIntSlice(hs.Squeeze(node, ballsMatcher, hs.ExtractChildrenTexts))
		if len(balls) != 7 {
			continue
		}
		drafts = append(drafts, Draft{
			Date:      drawDate,
			Balls:     balls[:6],
			ExtraBall: balls[6],
		})
	}
	return drafts
}

func toIntSlice(values []string) []int {
	numbers := make([]int, 0)
	for _, val := range values {
		v, err := strconv.Atoi(val)
		if err == nil {
			numbers = append(numbers, v)
		}
	}
	return numbers
}

func parseDate(rawDate string) *time.Time {
	dateFields := datePat.FindStringSubmatch(rawDate)
	timeFields := timePat.FindStringSubmatch(rawDate)
	month := dateFields[1]
	day := dateFields[2]
	year := dateFields[3]
	hour := timeFields[1]
	minute := timeFields[2]
	phase := timeFields[3]
	dateStr := fmt.Sprintf("%s %s %s %s:%s %s", year, month, day, hour, minute, phase)
	parsed, err := time.Parse(dateInputFmt, dateStr)
	if err != nil {
		return nil
	}
	return &parsed
}

func getDocument(url, agent string) (*html.Node, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare GET %s: %v", url, err)
	}
	req.Header.Set("User-Agent", agent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform GET %s: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse HTML of %s: %v", url, err)
	}
	return body, nil
}
