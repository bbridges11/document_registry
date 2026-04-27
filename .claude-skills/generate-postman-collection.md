---
name: generate-postman-collection
description: "Generate a Postman collection from Echo routes registered in RegisterRoutes(e *echo.Echo) functions under internal/adapters/**/handler.go"
---

# Generate Postman Collection from Echo Routes

You are helping generate a Postman collection for a Go service that uses the Echo HTTP framework.

## Goal

Inspect the codebase and generate a Postman collection JSON file from all routes registered through Echo in functions with the exact signature/pattern:

`RegisterRoutes(e *echo.Echo)`

These functions live in:

- `internal/adapters/**/handler.go`

The output should be a valid **Postman Collection v2.1** JSON file that can be imported directly into Postman.

## What to inspect

1. Search recursively under:
   - `internal/adapters/**/handler.go`

2. Find every function named:
   - `RegisterRoutes(e *echo.Echo)`

3. Extract all route registrations from those functions and any helper functions they call within the same package when needed.

4. Detect:
   - HTTP method
   - path
   - route group prefixes
   - tags/grouping candidates from folder/package names
   - request body type if obvious
   - path params
   - query params if obvious from handler logic
   - auth requirements if middleware or comments indicate it
   - summary/description from comments when available

## Route patterns to recognize

Support the common Echo registration styles:

- `e.GET("/path", h.GetThing)`
- `e.POST("/path", h.CreateThing)`
- `e.PUT("/path", h.UpdateThing)`
- `e.PATCH("/path", h.PatchThing)`
- `e.DELETE("/path", h.DeleteThing)`

Also support grouped routes such as:

- `g := e.Group("/v1")`
- `admin := g.Group("/admin")`
- `g.GET("/users", h.ListUsers)`

You must resolve full paths by composing nested group prefixes.

Also support routes registered through variables, for example:

- `api := e.Group("/api")`
- `v1 := api.Group("/v1")`
- `v1.GET("/documents", h.ListDocuments)`

## Output requirements

Generate:

1. A Postman collection file, typically:
   - `postman/<service-name>.postman_collection.json`

2. Optionally, when enough information exists, also generate:
   - `postman/local.postman_environment.json`

If the service name is not obvious, use:
- `app.postman_collection.json`

## Collection structure rules

### Collection info
- Use Postman Collection schema v2.1
- Put a sensible collection name based on module or repo name
- Include a description noting it was generated from Echo route registrations

### Folder organization
Prefer grouping routes by bounded context / adapter folder. Example:

- `documents`
- `document_versions`
- `approvals`
- `health`

If a package/folder name is ambiguous, group by first stable path segment.

### Requests
For each request include:

- name
- method
- full path using `{{base_url}}`
- path broken into segments
- headers
- query params when known
- request body example when inferred
- description

### Descriptions
For each request description include:
- handler function name
- source file path
- auth notes if inferred
- path params
- query params
- TODO markers where inference was uncertain

Example description format:

Handler: `GetDocument`  
Source: `internal/adapters/documents/handler.go`  
Auth: Bearer token required  
Path Params:
- `id`: document identifier

### Path variables
Convert Echo params like:
- `:id`
- `:documentID`

into Postman path variables:
- `{{id}}`
- `{{documentID}}`

Example:
- Echo route: `/documents/:id/versions/:versionID`
- Postman raw URL: `{{base_url}}/documents/{{id}}/versions/{{versionID}}`

### Headers
Default headers:
- `Accept: application/json`

For methods with JSON bodies, also include:
- `Content-Type: application/json`

If auth is inferred, include:
- `Authorization: Bearer {{token}}`

## Inferring request bodies

Try to infer request body shape from handler logic.

Look for patterns like:
- `var req CreateDocumentRequest`
- `req := new(CreateDocumentRequest)`
- `c.Bind(&req)`
- `c.Bind(req)`

Then inspect the struct definition and generate an example JSON body.

Rules:
- Include exported fields
- Use json tags when present
- Use representative placeholder values
- Omit fields tagged with `json:"-"`
- If nested structs exist, recurse reasonably
- If type inference is hard, create a minimal placeholder object and note uncertainty

Example:

```json
{
  "title": "Example Document",
  "type": "policy",
  "content": "Sample content"
}