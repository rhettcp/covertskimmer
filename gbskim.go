package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const slackGbLink = "https://hooks.slack.com/services/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"

type GBSkimmer struct {
	config GBConfig
}

type GBConfig struct {
	filters []string
}

type GBListing struct {
	title string
	link  string
}

func (gbl *GBListing) FormSlackMessage() (string, string, string) {
	return slackGbLink, gbl.title, gbl.link
}

func (gb *GBSkimmer) formUrl(keywords string) string {
	return fmt.Sprintf("https://www.gunbroker.com/All/search?Keywords=%s", url.QueryEscape(keywords))
}

func (gb *GBSkimmer) FindListings() []Listing {
	listings := make([]Listing, 0)

	for _, f := range gb.config.filters {
		res, _ := gb.skimListings(f)
		for _, r := range res {
			listings = append(listings, r)
		}
	}
	return listings
}

func (gb *GBSkimmer) skimListings(keyword string) ([]*GBListing, error) {
	url := gb.formUrl(keyword)
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

	results := make([]*GBListing, 0)
	// Find the review items
	doc.Find(".results-display-container").Children().Each(func(i int, s *goquery.Selection) {
		listing := &GBListing{}
		n := s.Find(".listing-text")
		n = n.Find("h4")
		n = n.Find("a")
		listing.title = strings.TrimSpace(n.Text())
		link, _ := n.Attr("href")
		listing.link = fmt.Sprintf("https://www.gunbroker.com/%s", link)
		results = append(results, listing)
	})
	return results, nil
}

/*
<div class="listing col-xs-6 col-sm-4 col-md-3 col-lg-3 col-xl-2 fake-tile">
<div class="highlighter    ">
<div class="listing-extra-info">
<div class="is-featured"><i class="glyphicon glyphicon-star"></i></div>
<a href="/item/897027682" target="_self" class="was-visited"><span class="glyphicon glyphicon-ok" title="Seen"></span></a>
<div class="close-btn glyphicon glyphicon-remove" role="button"></div>
<div class="quick-view-btn btn btn-primary" role="button">
<span>Quick View</span><span class="glyphicon glyphicon-search"></span>
</div>
</div>
<div class="listing-image">
<a href="/item/897027682" target="_self">
<img src="https://p1.gunbroker.com/pics/897027000/897027682/thumb.jpg" alt="E-Lander 7.62x39 10 Rd AR15/M/16 Magazine F-99913770">
</a>
</div>
 <div class="listing-text">
<h4><a href="/item/897027682" target="_self">E-Lander 7.62x39 10 Rd AR15/M/16 Magazine F-99913770</a></h4>
</div>
<div class="listing-figures">
<h5>Price</h5>
<a href="/item/897027682" target="_self" class="buy-now">$37.99</a>
</div>
<div class="listing-meta">
<div class="constant-meta">
<span class="item-number"><span class="hidden-xs hidden-sm hidden-md">Item</span> #<span class="hidden-xs hidden-sm hidden-md">:</span><a href="/item/897027682" target="_self">897027682</a></span>
<span class="time-left"><span>Qty: 1</span></span>
</div>
<div class="variable-meta">
<span class="buy-now-available"><i class="fa fa-dollar"></i>Buy Now Available</span>
<span class="immediate-checkout"><i class="fa fa-credit-card-alt"></i>Immediate Checkout</span>
<span class="cc-fees"><b>NO</b> Credit Card Fee</span>
</div>
</div>
<div class="listing-seller">
<span><a href="/a/feedback/profile/542226" target="_self">firearmsvet</a></span>
<span><a href="/a/feedback/profile/542226" target="_self">A+(1314)</a></span>
<ul>
<li>
<a href="https://support.gunbroker.com/hc/en-us/articles/225003687" target="_self" class="gb-badge gb-badge-sm verified-badge" title="verified"></a>
</li>
<li>
<a href="https://support.gunbroker.com/hc/en-us/articles/225003687" target="_self" class="gb-badge gb-badge-sm ffl-badge" title="ffl"></a>
</li>
</ul>
</div>
</div>
</div>
*/
