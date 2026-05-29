package providers

import "net/url"

type Filters struct {
	From              string   `mapstructure:"from"`
	To                string   `mapstructure:"to"`
	MatchStatusCodes  []string `mapstructure:"matchstatuscodes"`
	MatchMimeTypes    []string `mapstructure:"matchmimetypes"`
	FilterStatusCodes []string `mapstructure:"filterstatuscodes"`
	FilterMimeTypes   []string `mapstructure:"filtermimetypes"`
}

func (f *Filters) GetParameters(forWayback bool) string {
	form := url.Values{}
	if f.From != "" {
		form.Add("from", f.From)
	}

	if f.To != "" {
		form.Add("to", f.To)
	}

	switch forWayback {
	case true:
		// generate parameters for wayback
		if len(f.MatchMimeTypes) > 0 {
			for _, mt := range f.MatchMimeTypes {
				form.Add("filter", "mimetype:"+mt)
			}
		}

		if len(f.MatchStatusCodes) > 0 {
			for _, ms := range f.MatchStatusCodes {
				form.Add("filter", "statuscode:"+ms)
			}
		}

		if len(f.FilterStatusCodes) > 0 {
			for _, sc := range f.FilterStatusCodes {
				form.Add("filter", "!statuscode:"+sc)
			}
		}

		if len(f.FilterMimeTypes) > 0 {
			for _, mt := range f.FilterMimeTypes {
				form.Add("filter", "!mimetype:"+mt)
			}
		}
	default:
		// generate parameters for commoncrawl
		if len(f.MatchStatusCodes) > 0 {
			for _, ms := range f.MatchStatusCodes {
				form.Add("filter", "status:"+ms)
			}
		}

		if len(f.MatchMimeTypes) > 0 {
			for _, mt := range f.MatchMimeTypes {
				form.Add("filter", "mime:"+mt)
			}
		}

		if len(f.FilterStatusCodes) > 0 {
			for _, ms := range f.FilterStatusCodes {
				form.Add("filter", "!=status:"+ms)
			}
		}

		if len(f.FilterMimeTypes) > 0 {
			for _, fs := range f.FilterMimeTypes {
				form.Add("filter", "!=mime:"+fs)
			}
		}

	}

	params := form.Encode()
	if params != "" {
		return "&" + params
	}

	return params
}
