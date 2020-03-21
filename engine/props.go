package engine

import (
	"encoding/json"
	"net/url"
)

// Props : The scraping engine Properties and description about the engine (e.g NetNaijaEngine)
type Props struct {
	// Engine Interface
	Engine

	// Struct attributes
	Name        string
	BaseURL     *url.URL // The Base URL for the engine
	SearchURL   *url.URL // URL for searching
	ListURL     *url.URL // URL to return movie lists
	Description string
	mode        Mode // The mode of the operations (list, search)
}

// PropsJSON : JSON structure of all downloadable movies
type PropsJSON struct {
	Props
	BaseURL   string
	SearchURL string
	ListURL   string
}

// MarshalJSON Props structure to return from api
func (p *Props) MarshalJSON() ([]byte, error) {
	props := PropsJSON{
		Props:     *p,
		BaseURL:   p.BaseURL.String(),
		SearchURL: p.SearchURL.String(),
		ListURL:   p.ListURL.String(),
	}

	return json.Marshal(props)
}

func (p *Props) getParseURL() *url.URL {
	if p.mode == SearchMode {
		return p.SearchURL
	}
	return p.ListURL
}
