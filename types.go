package main

type Document struct {
	Version     string            `json:"swagger"`
	Host        string            `json:"host"`
	BasePath    string            `json:"basePath"`
	Schemes     []string          `json:"schemes"`
	Consumes    []string          `json:"consumes"`
	Produces    []string          `json:"produces"`
	Information Info              `json:"info"`
	Methods     map[string]Path   `json:"paths"`
	Schemas     map[string]Schema `json:"definitions"`
	Tags        []Tag             `json:"tags"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Path map[string]Operation

type Operation struct {
	Parameters  []Parameter         `json:"parameters"`
	Tags        []string            `json:"tags"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	ID          string              `json:"operationId"`
	Results     map[string]Response `json:"responses"`
}

type Response struct {
	Description string `json:"description,omitempty"`
	Schema      Schema `json:"schema,omitempty"`
}

type Position string

const (
	QueryPos    Position = "query"

	HeaderPos   Position = "header"
	PathPos     Position = "path"
	FormDataPos Position = "formData"
	BodyPos     Position = "body"
)

type Parameter struct {
	Name        string   `json:"name"`
	In          Position `json:"in"`
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Type        string   `json:"type"`
	Format      string   `json:"format"`
	EnumValues  []string `json:"enum,omitempty"`
}

type Schema struct {
	Format      string            `json:"format,omitempty"`
	Type        string            `json:"type,omitempty"`
	Properties  map[string]Schema `json:"properties,omitempty"`
	Ref         string            `json:"$ref,omitempty"`
	Description string            `json:"description,omitempty"`
	Items       *Schema           `json:"items,omitempty"`
	EnumValues  []string          `json:"enum,omitempty"`
}

type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
