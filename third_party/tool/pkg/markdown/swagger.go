package markdown

import (
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v2"
)

func DecodeJSON(data []byte) (*API, error) {
	var api API
	err := json.Unmarshal(data, &api)
	return &api, err
}

func DecodeYAML(data []byte) (*API, error) {
	var api API
	err := yaml.Unmarshal(data, &api)
	return &api, err
}

func FilterParameters(params []Parameter, t string) []Parameter {
	var results []Parameter
	for _, param := range params {
		if strings.ToLower(param.In) == t {
			results = append(results, param)
		}
	}
	return results
}

func FilterSchema(schema string) string {
	schema = strings.Replace(schema, "#/definitions/", "", 1)
	schema = strings.TrimPrefix(schema, "request.")
	return schema
}

func CollectSchema(definitions map[string]Definition, schema string) SchemaContext {
	schema = strings.Replace(schema, "#/definitions/", "", 1)
	return SchemaContext{
		schema,
		definitions,
		definitions[schema],
	}
}

func FormatAnchor(schema string) string {
	schema = strings.Replace(schema, "/:/g", "", 0)
	schemas := strings.Split(schema, " ")
	schema = strings.Join(schemas, "-")
	schema = strings.ToLower(schema)
	return schema
}

type APIError map[string]string

type API struct {
	Swagger     string                `json,yaml:"swagger"`
	Info        Info                  `json,yaml:"info"`
	Host        string                `json,yaml:"host"`
	BasePath    string                `json,yaml:"basePath"`
	Schemes     []string              `json,yaml:"schemes"`
	Consumes    []string              `json,yaml:"consumes"`
	Produces    []string              `json,yaml:"produces"`
	Paths       map[string]Methods    `json,yaml:"paths"`
	Definitions map[string]Definition `json,yaml:"definitions"`
}

type Methods map[string]*Operation

type Tag struct {
	Tag      string
	BasePath string
	Methods  []*Operation
}

type Operation struct {
	API         *API
	Operation   string
	Path        string
	Tags        []string              `json,yaml:"tags"`
	Description string                `json,yaml:"description"`
	OperationID string                `json,yaml:"operationId"`
	Summary     string                `json,yaml:"summary"`
	Parameters  []Parameter           `json,yaml:"parameters"`
	Responses   map[string]Response   `json,yaml:"responses"`
	Definitions map[string]Definition `json,yaml:"definitions"`
}

type Response struct {
	Description string       `json,yaml:"description"`
	Schema      SchemaObject `json,yaml:"schema"`
}

type SchemaObject struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Ref         string `json:"$ref"`
	Items       struct {
		Type string `json:"type"`
		Ref  string `json:"$ref"`
	} `json:"items"`
}

type Parameter struct {
	Name             string                 `json,yaml:"name"`
	In               string                 `json,yaml:"in"`
	Description      string                 `json,yaml:"description"`
	Required         bool                   `json,yaml:"required"`
	Type             string                 `json,yaml:"type"`
	CollectionFormat string                 `json,yaml:"collectionFormat"`
	Items            map[string]interface{} `json,yaml:"items"`
	Schema           SchemaObject           `json,yaml:"schema"`
}

type Info struct {
	Version        string  `json,yaml:"version"`
	Title          string  `json,yaml:"title"`
	Description    string  `json,yaml:"description"`
	TermsOfService string  `json,yaml:"termsOfService"`
	Contact        Contact `json,yaml:"contact"`
	License        License `json,yaml:"license"`
}

type Contact struct {
	Name  string `json,yaml:"name"`
	Email string `json,yaml:"email"`
	URL   string `json,yaml:"url"`
}

type License struct {
	Name string `json,yaml:"name"`
	URL  string `json,yaml:"url"`
}

type Definition struct {
	Type       string                  `json,yaml:"type"`
	Properties map[string]SchemaObject `json,yaml:"properties"`
}

type SchemaContext struct {
	TopRef      string
	Definitions map[string]Definition
	Definition  Definition
}
