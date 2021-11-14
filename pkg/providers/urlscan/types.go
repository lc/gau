package urlscan

import (
	"reflect"
	"strings"
)

var (
	_BaseURL = "https://urlscan.io/"
)

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
	for i := 0; i < len(sort); i++ {
		t := reflect.TypeOf(sort[i])
		v := reflect.ValueOf(sort[i])
		switch t.Kind() {
		case reflect.String:
			sortParam = append(sortParam, v.String())
		}
	}
	return strings.Join(sortParam, ",")
}
