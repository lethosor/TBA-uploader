package main

import (
	"fmt"
	"os"
	"strings"

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
	table := doc.Find("table").First()
	table.Find("*").RemoveClass(STRIP_CLASSES...)
	table.Find("*").RemoveAttr("data-bs-toggle")
	table.Find("img").RemoveAttr("src")
	table.RemoveAttr("class")

	html, err := goquery.OuterHtml(table)
	html = strings.Replace(html, "\t", "", -1)
	fmt.Print(html)
}
