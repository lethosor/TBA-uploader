package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

var STRIP_CLASSES = []string{
	"active",
	"col-sm-4",
	"fa-lg",
	"lead",
	"mb-3",
	"row",
	"text-center",
	"text-uppercase",
}

var STRIP_ATTRS = []string{
	"data-bs-toggle",
	"fill",
	"viewBox",
	"xmlns",
}

var STRIP_REGEXPS = []regexp.Regexp{
	*regexp.MustCompile("(?m)^\\s+"),
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("usage: %s FILENAME\n", os.Args[0])
		os.Exit(1)
	}
	filename := os.Args[1]

	reader, err := os.Open(filename)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}

	doc, err := goquery.NewDocumentFromReader(reader)
	table := doc.Find("table:has(*)").First()
	table.Find("svg path").Remove()
	table.Find("*").RemoveClass(STRIP_CLASSES...)
	for _, attr := range STRIP_ATTRS {
		table.Find("*").RemoveAttr(attr)
	}
	table.Find("img").RemoveAttr("src")
	table.RemoveAttr("class")

	html, err := goquery.OuterHtml(table)
	for _, re := range STRIP_REGEXPS {
		html = re.ReplaceAllString(html, "")
	}
	html += "\n"
	html = regexp.MustCompile("\n+").ReplaceAllString(html, "\n")
	fmt.Print(html)
}
