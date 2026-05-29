package commoncrawl

type apiResponse struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type paginationResult struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

type apiResult []struct {
	API string `json:"cdx-api"`
}
