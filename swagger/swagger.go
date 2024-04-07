package swagger

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/danielgtaylor/casing"
	"github.com/danielgtaylor/restish/cli"
	"github.com/gosimple/slug"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v2 "github.com/pb33f/libopenapi/datamodel/high/v2"
	"github.com/pb33f/libopenapi/utils"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

// reSwagger2 is a regex used to detect OpenAPI files from their contents.
var reSwagger2 = regexp.MustCompile(`['"]?swagger['"]?\s*:\s*['"]?2`)

// OpenAPI Extensions
const (
	// Change the CLI name for an operation or parameter
	ExtName = "x-cli-name"

	// Set additional command aliases for an operation
	ExtAliases = "x-cli-aliases"

	// Change the description of an operation or parameter
	ExtDescription = "x-cli-description"

	// Ignore a path, operation, or parameter
	ExtIgnore = "x-cli-ignore"

	// Create a hidden command for an operation. It will not show in the help,
	// but can still be called.
	ExtHidden = "x-cli-hidden"

	// Custom auto-configuration for CLIs
	ExtCLIConfig = "x-cli-config"
)

type autoConfig struct {
	Security string                       `json:"security"`
	Headers  map[string]string            `json:"headers,omitempty"`
	Prompt   map[string]cli.AutoConfigVar `json:"prompt,omitempty"`
	Params   map[string]string            `json:"params,omitempty"`
}

// Resolver is able to resolve relative URIs against a base.
type Resolver interface {
	GetBase() *url.URL
	Resolve(uri string) (*url.URL, error)
}

// getExt returns an extension converted to some type with the given default
// returned if the extension is not found or cannot be cast to that type.
func getExt[T any](v map[string]any, key string, def T) T {
	if v != nil {
		if i := v[key]; i != nil {
			if t, ok := i.(T); ok {
				return t
			}
		}
	}

	return def
}

// getExtSlice returns an extension converted to some type with the given
// default returned if the extension is not found or cannot be converted to
// a slice of the correct type.
func getExtSlice[T any](v map[string]any, key string, def []T) []T {
	if v != nil {
		if i := v[key]; i != nil {
			if s, ok := i.([]any); ok && len(s) > 0 {
				n := make([]T, len(s))
				for i := 0; i < len(s); i++ {
					if si, ok := s[i].(T); ok {
						n[i] = si
					}
				}
				return n
			}
		}
	}

	return def
}

// paramSchema returns a rendered schema line for a given parameter, falling
// back to the param type info if no schema is available.
func paramSchema(p *cli.Param, s *base.Schema) string {
	schemaDesc := fmt.Sprintf("(%s): %s", p.Type, p.Description)
	if s != nil {
		schemaDesc = renderSchema(s, "  ", modeWrite)
	}
	return schemaDesc
}

func openapiOperation(cmd *cobra.Command, method string, uriTemplate *url.URL, path *v2.PathItem, op *v2.Operation) cli.Operation {
	var pathParams, queryParams, headerParams []*cli.Param
	var pathSchemas, querySchemas, headerSchemas []*base.Schema = []*base.Schema{}, []*base.Schema{}, []*base.Schema{}

	// Combine path and operation parameters, with operation params having
	// precedence when there are name conflicts.
	combinedParams := []*v2.Parameter{}
	seen := map[string]bool{}
	for _, p := range op.Parameters {
		combinedParams = append(combinedParams, p)
		seen[p.Name] = true
	}
	for _, p := range path.Parameters {
		if !seen[p.Name] {
			combinedParams = append(combinedParams, p)
		}
	}

	for _, p := range combinedParams {
		if getExt(p.Extensions, ExtIgnore, false) {
			continue
		}

		var def interface{}
		var example interface{}

		typ := "string"
		var schema *base.Schema
		if p.Schema != nil && p.Schema.Schema() != nil {
			s := p.Schema.Schema()
			schema = s
			if len(s.Type) > 0 {
				// TODO: support params of multiple types?
				typ = s.Type[0]
			}

			if typ == "array" {
				if s.Items != nil && s.Items.IsA() {
					items := s.Items.A.Schema()
					if len(items.Type) > 0 {
						typ += "[" + items.Type[0] + "]"
					}
				}
			}

			def = s.Default
			example = s.Example
		}

		style := cli.StyleSimple

		displayName := getExt(p.Extensions, ExtName, "")
		description := getExt(p.Extensions, ExtDescription, p.Description)

		param := &cli.Param{
			Type:        typ,
			Name:        p.Name,
			DisplayName: displayName,
			Description: description,
			Style:       style,
			Default:     def,
			Example:     example,
		}

		switch p.In {
		case "path":
			if pathParams == nil {
				pathParams = []*cli.Param{}
			}
			pathParams = append(pathParams, param)
			pathSchemas = append(pathSchemas, schema)
		case "query":
			if queryParams == nil {
				queryParams = []*cli.Param{}
			}
			queryParams = append(queryParams, param)
			querySchemas = append(querySchemas, schema)
		case "header":
			if headerParams == nil {
				headerParams = []*cli.Param{}
			}
			headerParams = append(headerParams, param)
			headerSchemas = append(headerSchemas, schema)
		}
	}

	aliases := getExtSlice(op.Extensions, ExtAliases, []string{})

	name := casing.Kebab(op.OperationId)
	if name == "" {
		name = casing.Kebab(method + "-" + strings.Trim(uriTemplate.Path, "/"))
	}
	if override := getExt(op.Extensions, ExtName, ""); override != "" {
		name = override
	} else if oldName := slug.Make(op.OperationId); oldName != "" && oldName != name {
		// For backward-compatibility, add the old naming scheme as an alias
		// if it is different. See https://github.com/danielgtaylor/restish/issues/29
		// for additional context; we prefer kebab casing for readability.
		aliases = append(aliases, oldName)
	}

	desc := getExt(op.Extensions, ExtDescription, op.Description)
	hidden := getExt(op.Extensions, ExtHidden, false)

	if len(pathParams) > 0 {
		desc += "\n## Argument Schema:\n```schema\n{\n"
		for i, p := range pathParams {
			desc += "  " + p.OptionName() + ": " + paramSchema(p, pathSchemas[i]) + "\n"
		}
		desc += "}\n```\n"
	}

	if len(queryParams) > 0 || len(headerParams) > 0 {
		desc += "\n## Option Schema:\n```schema\n{\n"
		for i, p := range queryParams {
			desc += "  --" + p.OptionName() + ": " + paramSchema(p, querySchemas[i]) + "\n"
		}
		for i, p := range headerParams {
			desc += "  --" + p.OptionName() + ": " + paramSchema(p, headerSchemas[i]) + "\n"
		}
		desc += "}\n```\n"
	}

	mediaType := ""
	var examples []string

	codes := []string{}
	respMap := map[string]*v2.Response{}
	for k, v := range op.Responses.Codes {
		codes = append(codes, k)
		respMap[k] = v
	}

	sort.Strings(codes)

	type schemaEntry struct {
		code   string
		ct     string
		schema *base.Schema
	}
	schemaMap := map[[32]byte][]schemaEntry{}
	for _, code := range codes {
		// var resp *v2.Response
		if respMap[code] == nil {
			continue
		}

		// resp = respMap[code]

		hash := [32]byte{}

		if schemaMap[hash] == nil {
			schemaMap[hash] = []schemaEntry{}
		}
		schemaMap[hash] = append(schemaMap[hash], schemaEntry{
			code: code,
		})

	}

	schemaKeys := maps.Keys(schemaMap)
	sort.Slice(schemaKeys, func(i, j int) bool {
		return schemaMap[schemaKeys[i]][0].code < schemaMap[schemaKeys[j]][0].code
	})

	for _, s := range schemaKeys {
		entries := schemaMap[s]

		var resp *v2.Response
		if len(entries) == 1 && respMap[entries[0].code] != nil {
			resp = respMap[entries[0].code]
		}

		codeNums := []string{}
		for _, v := range entries {
			codeNums = append(codeNums, v.code)
		}

		hasSchema := s != [32]byte{}

		ct := ""
		if hasSchema {
			ct = " (" + entries[0].ct + ")"
		}

		if resp != nil {
			desc += "\n## Response " + entries[0].code + ct + "\n"
			respDesc := getExt(resp.Extensions, ExtDescription, resp.Description)
			if respDesc != "" {
				desc += "\n" + respDesc + "\n"
			} else if !hasSchema {
				desc += "\nResponse has no body\n"
			}
		} else {
			desc += "\n## Responses " + strings.Join(codeNums, "/") + ct + "\n"
			if !hasSchema {
				desc += "\nResponse has no body\n"
			}
		}

		headers := respMap[entries[0].code].Headers
		if len(headers) > 0 {
			keys := maps.Keys(headers)
			sort.Strings(keys)
			desc += "\nHeaders: " + strings.Join(keys, ", ") + "\n"
		}

		if hasSchema {
			desc += "\n```schema\n" + renderSchema(entries[0].schema, "", modeRead) + "\n```\n"
		}
	}

	tmpl := uriTemplate.String()
	if s, err := url.PathUnescape(uriTemplate.String()); err == nil {
		tmpl = s
	}

	// Try to add a group: if there's more than 1 tag, we'll just pick the
	// first one as a best guess
	group := ""
	if len(op.Tags) > 0 {
		group = op.Tags[0]
	}

	dep := ""

	return cli.Operation{
		Name:          name,
		Group:         group,
		Aliases:       aliases,
		Short:         op.Summary,
		Long:          strings.Trim(desc, "\n") + "\n",
		Method:        method,
		URITemplate:   tmpl,
		PathParams:    pathParams,
		QueryParams:   queryParams,
		HeaderParams:  headerParams,
		BodyMediaType: mediaType,
		Examples:      examples,
		Hidden:        hidden,
		Deprecated:    dep,
	}
}

func loadSwagger(cfg Resolver, cmd *cobra.Command, location *url.URL, resp *http.Response) (cli.API, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return cli.API{}, err
	}

	config := datamodel.NewOpenDocumentConfiguration()
	schemeLower := strings.ToLower(location.Scheme)
	if schemeLower == "http" || schemeLower == "https" {
		// Set the base URL to resolve relative references.
		config.BaseURL = &url.URL{Scheme: location.Scheme, Host: location.Host, Path: path.Dir(location.Path)}
	} else {
		// Set the base local directory path to resolve relative references.
		config.BasePath = path.Dir(location.Path)
	}

	doc, err := libopenapi.NewDocumentWithConfiguration(data, config)
	if err != nil {
		return cli.API{}, err
	}

	var v2Model v2.Swagger

	switch doc.GetSpecInfo().SpecType {
	case utils.OpenApi2:
		result, _ := doc.BuildV2Model()
		v2Model = result.Model
	default:
		return cli.API{}, fmt.Errorf("unsupported OpenAPI document")
	}

	// See if this server has any base path prefix we need to account for.
	basePath := cfg.GetBase().Path

	operations := []cli.Operation{}
	if v2Model.Paths != nil {
		for uri, path := range v2Model.Paths.PathItems {
			if getExt(path.Extensions, ExtIgnore, false) {
				continue
			}

			resolved, err := cfg.Resolve(strings.TrimSuffix(basePath, "/") + uri)
			if err != nil {
				return cli.API{}, err
			}

			for method, operation := range path.GetOperations() {
				if operation == nil || getExt(operation.Extensions, ExtIgnore, false) {
					continue
				}

				operations = append(operations, openapiOperation(cmd, strings.ToUpper(method), resolved, path, operation))
			}
		}
	}

	authSchemes := []cli.APIAuth{}

	short := ""
	long := ""
	if v2Model.Info != nil {
		short = getExt(v2Model.Info.Extensions, ExtName, v2Model.Info.Title)
		long = getExt(v2Model.Info.Extensions, ExtDescription, v2Model.Info.Description)
	}

	api := cli.API{
		Short:      short,
		Long:       long,
		Operations: operations,
	}

	if len(authSchemes) > 0 {
		api.Auth = authSchemes
	}

	loadAutoConfig(&api, &v2Model)

	return api, nil
}

func loadAutoConfig(api *cli.API, model *v2.Swagger) {
	var config *autoConfig

	cfg := model.Extensions[ExtCLIConfig]
	if cfg == nil {
		return
	}

	low := model.GoLow()
	for k, v := range low.Extensions {
		if k.Value == ExtCLIConfig {
			if err := v.ValueNode.Decode(&config); err != nil {
				fmt.Fprintf(os.Stderr, "Unable to unmarshal x-cli-config: %v", err)
				return
			}
			break
		}
	}

	authName := config.Security
	params := map[string]string{}

	// Params can override the values above if needed.
	for k, v := range config.Params {
		params[k] = v
	}

	api.AutoConfig = cli.AutoConfig{
		Headers: config.Headers,
		Prompt:  config.Prompt,
		Auth: cli.APIAuth{
			Name:   authName,
			Params: params,
		},
	}
}

type loader struct {
	location *url.URL
	base     *url.URL
}

func (l *loader) GetBase() *url.URL {
	return l.base
}

func (l *loader) Resolve(relURI string) (*url.URL, error) {
	parsed, err := url.Parse(relURI)
	if err != nil {
		return nil, err
	}

	return l.base.ResolveReference(parsed), nil
}

func (l *loader) LocationHints() []string {
	return []string{"/openapi.json", "/openapi.yaml", "openapi.json", "openapi.yaml"}
}

func (l *loader) Detect(resp *http.Response) bool {

	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return reSwagger2.Match(body)
}

func (l *loader) Load(entrypoint, spec url.URL, resp *http.Response) (cli.API, error) {
	l.location = &spec
	l.base = &entrypoint
	return loadSwagger(l, cli.Root, &spec, resp)
}

// New creates a new OpenAPI loader.
func New() cli.Loader {
	return &loader{}
}
