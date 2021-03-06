package openapi

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/danielgtaylor/restish/cli"
	"github.com/stretchr/testify/assert"
)

var sample = `
openapi: "3.0.0"
info:
  version: 1.0.0
  title: Swagger Petstore
  license:
    name: MIT
servers:
  - url: http://petstore.swagger.io/v1
paths:
  /pets:
    get:
      summary: List all pets
      operationId: listPets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: How many items to return at one time (max 100)
          required: false
          schema:
            type: integer
            format: int32
      responses:
        '200':
          description: A paged array of pets
          headers:
            x-next:
              description: A link to the next page of responses
              schema:
                type: string
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pets"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
    post:
      summary: Create a pet
      operationId: createPets
      tags:
        - pets
      responses:
        '201':
          description: Null response
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
  /pets/{petId}:
    get:
      summary: Info for a specific pet
      operationId: showPetById
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The id of the pet to retrieve
          schema:
            type: string
      responses:
        '200':
          description: Expected response to a valid request
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
        default:
          description: unexpected error
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
components:
  schemas:
    Pet:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      items:
        $ref: "#/components/schemas/Pet"
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
  securitySchemes:
    default:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://example.com/authorize
          tokenUrl: https://example.com/token
x-cli-config:
  security: default
  prompt:
    client_id:
      description: Client identifier
      example: abc123
`

func TestLoadOpenAPI(t *testing.T) {
	entry, _ := url.Parse("http://api.example.com")
	spec, _ := url.Parse("/openapi.yaml")

	resp := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(sample)),
	}

	api, err := New().Load(*entry, *spec, resp)
	assert.NoError(t, err)

	expected := cli.API{
		Short: "Swagger Petstore",
		Auth: []cli.APIAuth{
			{
				Name: "oauth-authorization-code",
				Params: map[string]string{
					"client_id":     "",
					"authorize_url": "https://example.com/authorize",
					"token_url":     "https://example.com/token",
				},
			},
		},
		Operations: []cli.Operation{
			{
				Name:         "createpets",
				Short:        "Create a pet",
				Long:         "\n## Response 201\n\nNull response\n\n## Response default (application/json)\n\nunexpected error\n\n```schema\n{\n  code*: (integer format:int32) \n  message*: (string) \n}\n```\n",
				Method:       "POST",
				URITemplate:  "http://api.example.com/pets",
				PathParams:   []*cli.Param{},
				QueryParams:  []*cli.Param{},
				HeaderParams: []*cli.Param{},
			},
			{
				Name:        "listpets",
				Short:       "List all pets",
				Long:        "\n## Response 200 (application/json)\n\nA paged array of pets\n\n```schema\n[\n  {\n    id*: (integer format:int64) \n    name*: (string) \n    tag: (string) \n  }\n]\n```\n\n## Response default (application/json)\n\nunexpected error\n\n```schema\n{\n  code*: (integer format:int32) \n  message*: (string) \n}\n```\n",
				Method:      "GET",
				URITemplate: "http://api.example.com/pets",
				PathParams:  []*cli.Param{},
				QueryParams: []*cli.Param{
					{
						Type:        "integer",
						Name:        "limit",
						Description: "How many items to return at one time (max 100)",
					},
				},
				HeaderParams: []*cli.Param{},
			},
			{
				Name:        "showpetbyid",
				Short:       "Info for a specific pet",
				Long:        "\n## Response 200 (application/json)\n\nExpected response to a valid request\n\n```schema\n{\n  id*: (integer format:int64) \n  name*: (string) \n  tag: (string) \n}\n```\n\n## Response default (application/json)\n\nunexpected error\n\n```schema\n{\n  code*: (integer format:int32) \n  message*: (string) \n}\n```\n",
				Method:      "GET",
				URITemplate: "http://api.example.com/pets/{petId}",
				PathParams: []*cli.Param{
					{
						Type:        "string",
						Name:        "petId",
						Description: "The id of the pet to retrieve",
					},
				},
				QueryParams:  []*cli.Param{},
				HeaderParams: []*cli.Param{},
			},
		},
		AutoConfig: cli.AutoConfig{
			Prompt: map[string]cli.AutoConfigVar{
				"client_id": {
					Description: "Client identifier",
					Example:     "abc123",
				},
			},
			Auth: cli.APIAuth{
				Name: "oauth-authorization-code",
				Params: map[string]string{
					"client_id":     "",
					"authorize_url": "https://example.com/authorize",
					"token_url":     "https://example.com/token",
				},
			},
		},
	}

	sort.Slice(api.Operations, func(i, j int) bool {
		return strings.Compare(api.Operations[i].Name, api.Operations[j].Name) < 0
	})

	assert.Equal(t, expected, api)
}
