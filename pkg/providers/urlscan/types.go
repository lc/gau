package urlscan

import (
	"strings"
	"fmt"
	jsoniter "github.com/json-iterator/go"
)

var _BaseURL = "https://urlscan.io/"

type apiResponse struct {
	Status  int            `json:"status"`
	Results []searchResult `json:"results"`
	HasMore bool           `json:"has_more"`
}

type searchResult struct {
	Page archivedPage
	Sort []interface{} `json:"sort"`
}

type archivedPage struct {
	Domain   string `json:"domain"`
	MimeType string `json:"mimeType"`
	URL      string `json:"url"`
	Status   string `json:"status"`
}

func parseSort(sort []interface{}) string {
    var sortParam []string

    for _, t := range sort {
        switch v := t.(type) {
        case string:
            sortParam = append(sortParam, v)
        case jsoniter.Number:
            sortParam = append(sortParam, v.String())
        default:
            sortParam = append(sortParam, fmt.Sprint(v))
        }
    }

    return strings.Join(sortParam, ",")
}
