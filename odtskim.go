package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const slackOdtLink = "https://hooks.slack.com/services/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"

var (
	urls = []OdtSearch{
		OdtSearch{
			Type:    "Handgun",
			Url:     "https://www.theoutdoorstrader.com/forums/handguns.72/",
			Filters: []string{"509", "iwi"},
		},
		OdtSearch{
			Type:    "Rifle",
			Url:     "https://www.theoutdoorstrader.com/forums/rifles.73/",
			Filters: []string{"iwi", "tavor"},
		},
		OdtSearch{
			Type:    "Shotgun",
			Url:     "https://www.theoutdoorstrader.com/forums/shotguns.258/",
			Filters: []string{"kalashnikov"},
		},
		OdtSearch{
			Type:    "Ammo",
			Url:     "https://www.theoutdoorstrader.com/forums/ammunition.70/",
			Filters: []string{"primer"},
		},
	}
)

type OdtSearch struct {
	Type    string
	Url     string
	Filters []string
}

type OdtListing struct {
	Title        string
	Link         string
	ForSale      bool
	ForTrade     bool
	Location     string
	Zipcode      string
	Price        string
	Bos          bool
	Description  string
	PictureLinks []string
}

func (o *OdtListing) FormSlackMessage() (string, string, string) {
	return slackOdtLink, o.Title, o.Link
}

type OdtSkimmer struct {
}

func (o *OdtSkimmer) FindListings() []Listing {
	allListings := make([]Listing, 0)
	for _, search := range urls {
		listings, _ := o.skimListings(search.Url, search.Filters)
		fmt.Println("Found:", len(listings))
		for _, l := range listings {
			allListings = append(allListings, l)
		}
	}

	fmt.Println("ODT Found:", len(allListings))
	return allListings
}

func (o *OdtSkimmer) extractListingData(listings *[]*OdtListing) error {
	for _, listing := range *listings {
		if strings.HasPrefix(listing.Title, "FS / FT") {
			listing.ForSale = true
			listing.ForTrade = true
		} else if strings.HasPrefix(listing.Title, "FS") {
			listing.ForSale = true
		} else if strings.HasPrefix(listing.Title, "FT") {
			listing.ForSale = true
		}

		res, err := http.Get(listing.Link)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			return err
		}

		// Load the HTML document
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return err
		}

		// Find the review items
		doc.Find(".listingSection").Each(func(i int, s *goquery.Selection) {
			itemtext := s.Text()
			locInd := strings.Index(itemtext, "Location:")
			itemtext = itemtext[:locInd+len("Location:")]
			brInd := strings.Index(itemtext, "\n")
			location := itemtext[brInd:]
			itemtext = itemtext[:brInd+1]
			listing.Location = location
		})
	}
	return nil
}

//skimListings hands the url for handguns, rifles, or shotguns
func (o *OdtSkimmer) skimListings(url string, matches []string) ([]*OdtListing, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	results := make([]*OdtListing, 0)
	// Find the review items
	doc.Find(".Listing").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s.Text())
		itemtext := s.Text()
		link, _ := s.Attr("href")
		match := false
		for _, m := range matches {
			if strings.Contains(strings.ToLower(itemtext), m) {
				match = true
				break
			}
		}
		if match {
			results = append(results, &OdtListing{Title: itemtext, Link: "https://www.theoutdoorstrader.com/" + link})
		}
	})

	err = o.extractListingData(&results)
	if err != nil {
		panic(err)
	}

	return results, nil
}
