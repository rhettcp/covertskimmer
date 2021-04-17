package main

import (
	"time"
)

type SkimmerEngine struct {
	config   SkimmerConfig
	skimmers []Skimmer
	memCache map[string]time.Time
	done     chan struct{}
}

type SkimmerConfig struct {
	SkimmerInterval time.Duration
}

type Skimmer interface {
	FindListings() []Listing
}

type Listing interface {
	FormSlackMessage() (string, string, string)
}

func NewSkimmerEngine(skimmerConfig SkimmerConfig) (*SkimmerEngine, error) {
	return &SkimmerEngine{
		config:   skimmerConfig,
		skimmers: make([]Skimmer, 0),
		memCache: make(map[string]time.Time),
		done:     make(chan struct{}),
	}, nil
}

func (se *SkimmerEngine) AddSkimmer(skimmer Skimmer) {
	se.skimmers = append(se.skimmers, skimmer)
}

func (se *SkimmerEngine) RunSkimmerEngine() {
	ticker := time.NewTicker(se.config.SkimmerInterval)
	se.runSkimmers()
	for {
		select {
		case <-ticker.C:
			se.runSkimmers()
		case <-se.done:
			return
		}
	}
}

func (se *SkimmerEngine) Stop() error {
	close(se.done)
	return nil
}

func (se *SkimmerEngine) runSkimmers() {
	for _, s := range se.skimmers {
		listings := s.FindListings()
		for _, l := range listings {
			slackLink, listTitle, listLink := l.FormSlackMessage()
			se.SlackOrForget(slackLink, listTitle, listLink)
		}
	}
}

func (se *SkimmerEngine) SlackOrForget(slackLink, title, link string) {
	if alertTime, ok := se.memCache[link]; !ok {
		//alert and add time.
		//SendSlackNotification(slackLink, title, link)
		se.memCache[link] = time.Now()
	} else {
		if time.Now().Add(48 * time.Hour).Before(alertTime) {
			delete(se.memCache, link)
		}
	}
}
