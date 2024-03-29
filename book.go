package book

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

const (
	BookTitlePrefix      = "Book title: "
	BookURLIndicator     = "/book/show/"
	BookIDIndicator      = "/work/quotes/"
	BookCoverIndicator   = "BookCover__image"
	BookAuthorsIndicator = "ContributorLinksList"
	BookGenresIndicator  = "/genres/"
	BookRatingIndicator  = "RatingStatistics__rating"
	BookStatsIndicator   = "RatingStatistics__meta"
)

type Books struct {
	Books []Book `json:"books"`
}

type Book struct {
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	ID       string   `json:"id"`
	CoverUrl string   `json:"cover_url"`
	Authors  []string `json:"authors"`
	Genres   []string `json:"genres"`
	Rating   float64  `json:"rating"`
	Ratings  int      `json:"ratings"`
	Reviews  int      `json:"reviews"`
}

func GetBookURLs(r io.Reader) ([]string, error) {
	bookURLs := []string{}

	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	extractURLs(doc, &bookURLs)

	return bookURLs, nil
}

func extractURLs(n *html.Node, urls *[]string) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				url := attr.Val

				if strings.HasPrefix(url, BookURLIndicator) {
					*urls = append(*urls, url)
				}

				break
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractURLs(c, urls)
	}
}

func GetBook(r io.Reader) (*Book, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	book := &Book{}

	extractBookInfo(doc, book)

	return book, nil
}

func extractBookInfo(n *html.Node, curBook *Book) {
	if n.Type == html.ElementNode && n.Data == "a" {
		extractID(n, curBook)
		extractGenres(n, curBook)
	}

	if n.Type == html.ElementNode && n.Data == "div" {
		extractCover(n, curBook)
		extractRating(n, curBook)
		extractStats(n, curBook)
		extractAuthors(n, curBook)
	}

	if n.Type == html.ElementNode && n.Data == "h1" {
		extractTitle(n, curBook)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractBookInfo(c, curBook)
	}
}

func extractRating(n *html.Node, curBook *Book) {
	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == BookRatingIndicator {
			textNode := n.FirstChild

			if textNode != nil {
				val, err := strconv.ParseFloat(textNode.Data, 64)
				if err != nil {
					fmt.Println(err)
				}

				curBook.Rating = val
			}

			break
		}
	}
}

func extractStats(n *html.Node, curBook *Book) {
	correctClass, val := false, ""

	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == BookStatsIndicator {
			correctClass = true
		}

		if attr.Key == "aria-label" {
			val = attr.Val
		}

		if correctClass && val != "" {
			break
		}
	}

	if correctClass {
		parts := strings.Split(val, " ")
		ratings := parts[0]
		reviews := parts[3]
		ratings = strings.Join(strings.Split(ratings, ","), "")
		reviews = strings.Join(strings.Split(reviews, ","), "")

		ratingsVal, err := strconv.Atoi(ratings)
		if err != nil {
			fmt.Println(err)
		}

		curBook.Ratings = ratingsVal

		reviewsVal, err := strconv.Atoi(reviews)
		if err != nil {
			fmt.Println(err)
		}

		curBook.Reviews = reviewsVal
	}
}

func extractGenres(n *html.Node, curBook *Book) {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			url := attr.Val

			if strings.Contains(url, BookGenresIndicator) {
				parts := strings.Split(url, "/")
				genre := parts[len(parts)-1]
				curBook.Genres = append(curBook.Genres, genre)
			}

			break
		}
	}
}

func extractAuthors(n *html.Node, curBook *Book) {
	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == BookAuthorsIndicator {
			authors := []string{}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				aNode := c.FirstChild
				if aNode == nil || aNode.Data != "a" {
					continue
				}

				spanNode := aNode.FirstChild
				if spanNode == nil || spanNode.Data != "span" {
					continue
				}

				name := spanNode.FirstChild
				if name.Type != html.TextNode {
					continue
				}

				authors = append(authors, name.Data)
			}

			curBook.Authors = authors
			break
		}
	}
}

func extractCover(n *html.Node, curBook *Book) {
	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == BookCoverIndicator {
			targetDiv := n.FirstChild
			if targetDiv == nil {
				continue
			}

			imageNode := targetDiv.FirstChild
			if imageNode == nil || imageNode.Data != "img" {
				continue
			}

			correctClass, correctRole, imgSRC := false, false, ""

			for _, attr := range imageNode.Attr {
				if attr.Key == "class" && attr.Val == "ResponsiveImage" {
					correctClass = true
				}

				if attr.Key == "role" && attr.Val == "presentation" {
					correctRole = true
				}

				if attr.Key == "src" {
					imgSRC = attr.Val
				}

				if correctClass && correctRole && imgSRC != "" {
					break
				}
			}

			if correctClass && correctRole {
				curBook.CoverUrl = imgSRC
			}
		}
	}
}

func extractID(n *html.Node, curBook *Book) {
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			url := attr.Val

			if strings.Contains(url, BookIDIndicator) {
				parts := strings.Split(url, "/")
				id := parts[len(parts)-1]
				curBook.ID = id
			}

			break
		}
	}
}

func extractTitle(n *html.Node, curBook *Book) {
	correctClass, correctData, title := false, false, ""

	for _, attr := range n.Attr {
		if attr.Key == "class" && attr.Val == "Text Text__title1" {
			correctClass = true
		}

		if attr.Key == "data-testid" && attr.Val == "bookTitle" {
			correctData = true
		}

		if attr.Key == "aria-label" {
			title = attr.Val[len(BookTitlePrefix):]
		}

		if correctClass && correctData && title != "" {
			break
		}
	}

	if correctClass && correctData {
		curBook.Title = title
	}
}
