package upguard

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/danielgtaylor/restish/cli"
	"github.com/spf13/viper"
	// "github.com/google/martian/log"
)

// UpguardAPIParser parses JSON:API hypermedia links.
type UpguardAPIParser struct{}

var rePageToken = regexp.MustCompile(`page_token=(\d+)`)

// var reOpenAPI3 = regexp.MustCompile(`['"]?openapi['"]?\s*:\s*['"]?3`)

// ParseLinks processes the links in a parsed response.
func (j UpguardAPIParser) ParseLinks(base *url.URL, resp *cli.Response) error {
	if !viper.GetBool("rsh-no-paginate") {
		if b, ok := resp.Body.(map[string]interface{}); ok {
			// resp might be paginated if total_results is set
			if b["total_results"] != nil {
				cli.LogDebug("[UpguardAPIParser] Found a potential pagination situation (total_results=%i)", b["total_results"])
				// resp is paginated if next_page_token is set
				if b["next_page_token"] != nil {
					rawQuery := base.RawQuery
					if rePageToken.MatchString(rawQuery) {
						rawQuery = rePageToken.ReplaceAllString(rawQuery, "page_token="+b["next_page_token"].(string))
					} else {
						rawQuery = fmt.Sprintf("page_token=%s&%s", b["next_page_token"].(string), rawQuery)
					}
					cli.LogDebug("[UpguardAPIParser] rawQuery: %s", rawQuery)

					rel := "next"
					resp.Links[rel] = append(resp.Links[rel], &cli.Link{
						Rel: rel,
						URI: fmt.Sprintf("?%s", rawQuery),
					})
				}
				for key := range b {
					if _, ok := b[key].([]interface{}); ok {
						// resp.Body = b["domains"]
						cli.LogDebug("[UpguardAPIParser] Found a key " + key)
						resp.Body = b[key]
						break
					} else {
						cli.LogDebug("[UpguardAPIParser] Not a usable key " + key)
					}
				}
			}
		}
	}
	return nil
}
