package output

import (
	"bufio"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type JSONResult struct {
	Url string `json:"url"`
}

func WriteURLs(results <-chan string, writer io.Writer) error {
	wr := bufio.NewWriter(writer)
	str := &strings.Builder{}
	for result := range results {
		str.WriteString(result)
		str.WriteRune('\n')
		_, err := wr.WriteString(str.String())
		if err != nil {
			wr.Flush()
			return err
		}
		str.Reset()
	}
	return wr.Flush()
}
func WriteURLsJSON(results <-chan string, writer io.Writer) {
	var jr JSONResult
	enc := jsoniter.NewEncoder(writer)
	for result := range results {
		jr.Url = result
		enc.Encode(jr)
	}
}
