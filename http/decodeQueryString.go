package http

import (
	"net/url"

	"github.com/gorilla/schema"
)

var (
	queryDecoder = func() *schema.Decoder {
		decoder := schema.NewDecoder()
		decoder.SetAliasTag("query_string")
		decoder.IgnoreUnknownKeys(true)
		return decoder
	}()
)

// DecodeQueryParams decode a query url using json tag from struct
func DecodeQueryParams(src url.Values, dst interface{}) error {
	return queryDecoder.Decode(dst, src)
}
