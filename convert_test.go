package conv_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	conv "github.com/duh-rpc/openapi-markdown.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertMinimalMarkdown(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  string
		wantErr string
	}{
		{
			name: "title and description",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  description: Test Description
  version: 1.0.0
paths: {}`,
			opts: conv.ConvertOptions{
				Title:       "Test API",
				Description: "Test Description",
			},
			wantMd: "# Test API\n\nTest Description\n\n",
		},
		{
			name: "title only",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: "# Test API\n\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantMd, string(result.Markdown))
			assert.Equal(t, 0, result.EndpointCount)
			assert.Equal(t, 0, result.TagCount)
		})
	}
}

func TestConvertEmptyInput(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantErr string
	}{
		{
			name:    "empty openapi bytes",
			openapi: "",
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantErr: "openapi input cannot be empty",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.ErrorContains(t, err, test.wantErr)
			require.Nil(t, result)
		})
	}
}

func TestConvertEmptyTitle(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantErr string
	}{
		{
			name: "empty title",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			opts: conv.ConvertOptions{
				Title: "",
			},
			wantErr: "title cannot be empty",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.ErrorContains(t, err, test.wantErr)
			require.Nil(t, result)
		})
	}
}

func TestConvertUnsupportedVersion(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantErr string
	}{
		{
			name: "openapi 2.0",
			openapi: `swagger: "2.0"
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantErr: "only openapi 3.x is supported",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.ErrorContains(t, err, test.wantErr)
			require.Nil(t, result)
		})
	}
}

func TestConvertTableOfContents(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "single endpoint",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## Table of Contents",
				"HTTP Request | Description",
				"-------------|------------",
				"GET [/users](#getusers) | List users",
			},
		},
		{
			name: "multiple endpoints",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
    post:
      summary: Create user
  /users/{id}:
    get:
      summary: Get user`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## Table of Contents",
				"GET [/users](#getusers) | List users",
				"POST [/users](#postusers) | Create user",
				"GET [/users/{id}](#getusersid) | Get user",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertSingleEndpoint(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "endpoint with description",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      description: Returns a list of all users`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## GET /users",
				"Returns a list of all users",
			},
		},
		{
			name: "endpoint with summary only",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## GET /users",
				"List users",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertMultipleEndpoints(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "multiple paths",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
  /posts:
    get:
      summary: List posts`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## GET /users",
				"List users",
				"## GET /posts",
				"List posts",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			assert.Equal(t, 2, result.EndpointCount)
		})
	}
}

func TestConvertUntaggedOperations(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "untagged operation",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## GET /users",
				"List users",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertMultipleTagsPerOperation(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "operation with multiple tags",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      tags:
        - Users
        - Admin`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## Admin",
				"### GET /users",
				"## Users",
				"### GET /users",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertPathParameters(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "path parameter",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      summary: Get user
      parameters:
        - name: id
          in: path
          required: true
          description: User ID
          schema:
            type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Path Parameters",
				"`id` (string, required)",
				"User ID",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertQueryParameters(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "query parameter",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: limit
          in: query
          required: false
          description: Maximum number of results
          schema:
            type: integer`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Query Parameters",
				"**limit** (integer)",
				"Maximum number of results",
			},
		},
		{
			name: "multiple query parameters",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: limit
          in: query
          required: false
          description: Maximum number of results
          schema:
            type: integer
        - name: offset
          in: query
          required: false
          description: Number of results to skip
          schema:
            type: integer`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Query Parameters",
				"**limit** (integer)",
				"Maximum number of results",
				"**offset** (integer)",
				"Number of results to skip",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertHeaderParameters(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "header parameter",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: X-API-Key
          in: header
          required: true
          description: API authentication key
          schema:
            type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Headers",
				"Name | Description | Required | Type",
				"-----|-------------|----------|-----",
				"X-API-Key | API authentication key | true | string",
			},
		},
		{
			name: "multiple header parameters",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: X-API-Key
          in: header
          required: true
          description: API authentication key
          schema:
            type: string
        - name: X-Request-ID
          in: header
          required: false
          description: Request tracking ID
          schema:
            type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Headers",
				"X-API-Key | API authentication key | true | string",
				"X-Request-ID | Request tracking ID | false | string",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertResponseExamples(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "response with schema reference",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Responses",
				"#### 200 Response",
				"Success",
				"```json",
				"\"id\":",
				"\"name\":",
				"```",
			},
		},
		{
			name: "multiple response codes",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: Not found
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Responses",
				"#### 200 Response",
				"Success",
				"#### 404 Response",
				"Not found",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertCompleteEndpoint(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "endpoint with all features",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      summary: Get user
      description: Returns a single user by ID
      parameters:
        - name: id
          in: path
          required: true
          description: User ID
          schema:
            type: string
        - name: fields
          in: query
          required: false
          description: Fields to include
          schema:
            type: string
        - name: X-API-Key
          in: header
          required: true
          description: API key
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## GET /users/{id}",
				"Returns a single user by ID",
				"#### Path Parameters",
				"`id` (string, required)",
				"User ID",
				"#### Query Parameters",
				"**fields** (string)",
				"Fields to include",
				"#### Headers",
				"X-API-Key | API key | true | string",
				"### Responses",
				"#### 200 Response",
				"Success",
				"#### 404 Response",
				"User not found",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			assert.Equal(t, 1, result.EndpointCount)
			assert.Equal(t, 1, result.TagCount)
		})
	}
}

func TestConvertDebugMode(t *testing.T) {
	for _, test := range []struct {
		name      string
		openapi   string
		opts      conv.ConvertOptions
		wantDebug func(*testing.T, *conv.DebugInfo)
	}{
		{
			name: "debug info collected",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      tags:
        - Users
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        '200':
          description: Success
  /posts:
    get:
      summary: List posts
      responses:
        '200':
          description: Success
        '404':
          description: Not found`,
			opts: conv.ConvertOptions{
				Title: "Test API",
				Debug: true,
			},
			wantDebug: func(t *testing.T, debug *conv.DebugInfo) {
				require.NotNil(t, debug)
				assert.Equal(t, 2, debug.ParsedPaths)
				assert.Equal(t, 2, debug.ExtractedOps)
				assert.Equal(t, 1, debug.UntaggedOps)
				assert.Contains(t, debug.TagsFound, "Default APIs")
				assert.Contains(t, debug.TagsFound, "Users")
				assert.Equal(t, 1, debug.ParameterCounts["query"])
				assert.Equal(t, 2, debug.ResponseCounts["200"])
				assert.Equal(t, 1, debug.ResponseCounts["404"])
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			test.wantDebug(t, result.Debug)
		})
	}
}

func TestConvertIntegrationPetStore(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "petstore endpoints",
			openapi: `openapi: 3.0.0
info:
  title: Petstore API
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List all pets
      tags:
        - Pets
      parameters:
        - name: limit
          in: query
          description: How many items to return
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: A list of pets
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Pets'
    post:
      summary: Create a pet
      tags:
        - Pets
      responses:
        '201':
          description: Pet created
  /pets/{petId}:
    get:
      summary: Info for a specific pet
      tags:
        - Pets
      parameters:
        - name: petId
          in: path
          required: true
          description: The id of the pet
          schema:
            type: string
      responses:
        '200':
          description: Expected pet
components:
  schemas:
    Pets:
      type: array
      items:
        $ref: '#/components/schemas/Pet'
    Pet:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Petstore API",
			},
			wantMd: []string{
				"# Petstore API",
				"## Table of Contents",
				"GET [/pets](#getpets) | List all pets",
				"POST [/pets](#postpets) | Create a pet",
				"GET [/pets/{petId}](#getpetspetid) | Info for a specific pet",
				"## GET /pets",
				"List all pets",
				"#### Query Parameters",
				"**limit** (integer)",
				"How many items to return",
				"### Responses",
				"#### 200 Response",
				"## POST /pets",
				"Create a pet",
				"### Responses",
				"#### 201 Response",
				"## GET /pets/{petId}",
				"#### Path Parameters",
				"`petId` (string, required)",
				"The id of the pet",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			assert.Equal(t, 3, result.EndpointCount)
			assert.Equal(t, 1, result.TagCount)
		})
	}
}

func TestConvertIntegrationComplexAPI(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "api with multiple tags",
			openapi: `openapi: 3.0.0
info:
  title: Complex API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      tags:
        - Users
      responses:
        '200':
          description: Success
    post:
      summary: Create user
      tags:
        - Users
      responses:
        '201':
          description: Created
  /posts:
    get:
      summary: List posts
      tags:
        - Posts
      responses:
        '200':
          description: Success
  /comments:
    get:
      summary: List comments
      tags:
        - Comments
      responses:
        '200':
          description: Success`,
			opts: conv.ConvertOptions{
				Title: "Complex API",
			},
			wantMd: []string{
				"# Complex API",
				"## Table of Contents",
				"## Comments",
				"### GET /comments",
				"## Posts",
				"### GET /posts",
				"## Users",
				"### GET /users",
				"### POST /users",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			assert.Equal(t, 4, result.EndpointCount)
			assert.Equal(t, 3, result.TagCount)
		})
	}
}

func TestConvertResponseInlineSchemaError(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantErr string
	}{
		{
			name: "inline schema not allowed",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantErr: "inline schemas not supported",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.ErrorContains(t, err, test.wantErr)
			require.Nil(t, result)
		})
	}
}

func TestConvertSingleTagOmitted(t *testing.T) {
	for _, test := range []struct {
		name      string
		openapi   string
		opts      conv.ConvertOptions
		wantMd    []string
		notWantMd []string
	}{
		{
			name: "single tag heading omitted",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      tags:
        - Users
  /users/{id}:
    get:
      summary: Get user
      tags:
        - Users`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## GET /users",
				"## GET /users/{id}",
			},
			notWantMd: []string{
				"## Users",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			for _, notWant := range test.notWantMd {
				assert.NotContains(t, md, notWant)
			}
		})
	}
}

func TestConvertDefaultAPIsLastPosition(t *testing.T) {
	for _, test := range []struct {
		name       string
		openapi    string
		opts       conv.ConvertOptions
		checkOrder func(*testing.T, string)
	}{
		{
			name: "default apis sorted last",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /untagged:
    get:
      summary: Untagged endpoint
  /users:
    get:
      summary: List users
      tags:
        - Users
  /posts:
    get:
      summary: List posts
      tags:
        - Posts`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			checkOrder: func(t *testing.T, md string) {
				postsIdx := strings.Index(md, "## Posts")
				usersIdx := strings.Index(md, "## Users")
				defaultIdx := strings.Index(md, "## Default APIs")

				assert.Greater(t, defaultIdx, postsIdx)
				assert.Greater(t, defaultIdx, usersIdx)
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			test.checkOrder(t, md)
		})
	}
}

func TestConvertCompleteExample(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "complete example with tags",
			openapi: `openapi: 3.0.0
info:
  title: Complete API
  description: A complete example API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      description: Returns a list of users
      tags:
        - Users
      parameters:
        - name: limit
          in: query
          required: false
          description: Maximum number of users
          schema:
            type: integer
        - name: X-API-Key
          in: header
          required: true
          description: API key
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserList'
        '400':
          description: Bad request
        '401':
          description: Unauthorized
  /users/{id}:
    get:
      summary: Get user
      description: Returns a specific user
      tags:
        - Users
      parameters:
        - name: id
          in: path
          required: true
          description: User ID
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
  /posts:
    get:
      summary: List posts
      tags:
        - Posts
      responses:
        '200':
          description: Success
components:
  schemas:
    UserList:
      type: array
      items:
        $ref: '#/components/schemas/User'
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        email:
          type: string`,
			opts: conv.ConvertOptions{
				Title:       "Complete API",
				Description: "A complete example API",
			},
			wantMd: []string{
				"# Complete API",
				"A complete example API",
				"## Table of Contents",
				"GET [/users](#getusers) | List users",
				"GET [/users/{id}](#getusersid) | Get user",
				"GET [/posts](#getposts) | List posts",
				"## Posts",
				"### GET /posts",
				"## Users",
				"### GET /users",
				"Returns a list of users",
				"#### Query Parameters",
				"**limit** (integer)",
				"Maximum number of users",
				"#### Headers",
				"X-API-Key | API key | true | string",
				"### Responses",
				"#### 200 Response",
				"#### 400 Response",
				"#### 401 Response",
				"### GET /users/{id}",
				"Returns a specific user",
				"#### Path Parameters",
				"`id` (string, required)",
				"User ID",
				"### Responses",
				"#### 404 Response",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			assert.Equal(t, 3, result.EndpointCount)
			assert.Equal(t, 2, result.TagCount)
		})
	}
}

func TestCompleteExampleWithGoldenFile(t *testing.T) {
	openapi, err := os.ReadFile("examples/openapi.yaml")
	require.NoError(t, err)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title:       "Pet Store API",
		Description: "A comprehensive API for managing a pet store with users, pets, and orders",
	})
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/golden/petstore-example.md")
	require.NoError(t, err)

	if !bytes.Equal(result.Markdown, expected) {
		actualPath := "testdata/golden/petstore-example.actual.md"
		err := os.WriteFile(actualPath, result.Markdown, 0644)
		require.NoError(t, err)

		t.Fatalf("Output doesn't match golden file\nActual written to: %s\nRun: diff %s testdata/golden/petstore-example.md",
			actualPath, actualPath)
	}
}

// Phase 1 tests: Request body rendering

func TestConvertRequestBodyPOST(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "POST with request body",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUser'
      responses:
        '201':
          description: Created
components:
  schemas:
    CreateUser:
      type: object
      required:
        - name
        - email
      properties:
        name:
          type: string
          description: User's full name
        email:
          type: string
          description: User's email address
        age:
          type: integer
          description: User's age`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Request",
				"```json",
				"\"name\":",
				"\"email\":",
				"```",
				"#### Field Definitions",
				"**name** (string, required)",
				"User's full name",
				"**email** (string, required)",
				"User's email address",
				"**age** (integer)",
				"User's age",
			},
		},
		{
			name: "PUT with request body",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    put:
      summary: Update user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUser'
      responses:
        '200':
          description: Updated
components:
  schemas:
    UpdateUser:
      type: object
      properties:
        name:
          type: string
          description: Updated name`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Request",
				"```json",
				"```",
				"#### Field Definitions",
				"**name** (string)",
				"Updated name",
			},
		},
		{
			name: "PATCH with request body",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    patch:
      summary: Patch user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PatchUser'
      responses:
        '200':
          description: Patched
components:
  schemas:
    PatchUser:
      type: object
      properties:
        status:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Request",
				"#### Field Definitions",
				"**status** (string)",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertRequestBodyGETSkipped(t *testing.T) {
	for _, test := range []struct {
		name      string
		openapi   string
		opts      conv.ConvertOptions
		notWantMd []string
	}{
		{
			name: "GET without request body",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			notWantMd: []string{
				"### Request",
				"#### Field Definitions",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, notWant := range test.notWantMd {
				assert.NotContains(t, md, notWant)
			}
		})
	}
}

func TestConvertRequestBodyArrayField(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "request with array field",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /batch:
    post:
      summary: Batch create
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/BatchRequest'
      responses:
        '201':
          description: Created
components:
  schemas:
    BatchRequest:
      type: object
      required:
        - items
      properties:
        items:
          type: array
          items:
            type: string
          description: List of item IDs`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Request",
				"#### Field Definitions",
				"**items** (string array, required)",
				"List of item IDs",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertRequestBodyEnumField(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "request with enum field",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUser'
      responses:
        '201':
          description: Created
components:
  schemas:
    CreateUser:
      type: object
      required:
        - role
      properties:
        role:
          type: string
          description: User role
          enum:
            - ADMIN
            - USER
            - GUEST`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Request",
				"#### Field Definitions",
				"**role** (string, required)",
				"User role Enums: `ADMIN`, `USER`, `GUEST`",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertFieldDefinitionsNested(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "nested object creates separate definition section",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /resource:
    post:
      summary: Create resource
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Resource'
      responses:
        '201':
          description: Created
components:
  schemas:
    Resource:
      type: object
      required:
        - name
        - metadata
      properties:
        name:
          type: string
          description: Resource name
        metadata:
          $ref: '#/components/schemas/Metadata'
    Metadata:
      type: object
      required:
        - created
      properties:
        created:
          type: string
          description: Creation timestamp
        tags:
          type: array
          items:
            type: string
          description: Resource tags`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Field Definitions",
				"**name** (string, required)",
				"Resource name",
				"**metadata** (object, required)",
				"**Metadata**",
				"- `created` (string, required): Creation timestamp",
				"- `tags` (string array): Resource tags",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertFieldDefinitionsDeeplyNested(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "deeply nested objects (3 levels)",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /resource:
    post:
      summary: Create resource
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Level1'
      responses:
        '201':
          description: Created
components:
  schemas:
    Level1:
      type: object
      properties:
        name:
          type: string
        level2:
          $ref: '#/components/schemas/Level2'
    Level2:
      type: object
      properties:
        value:
          type: string
        level3:
          $ref: '#/components/schemas/Level3'
    Level3:
      type: object
      properties:
        deep:
          type: string
          description: Deeply nested value`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"**Level2**",
				"- `value` (string)",
				"- `level3` (object)",
				"**Level3**",
				"- `deep` (string): Deeply nested value",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertFieldDefinitionsRecursive(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "recursive schema capped at depth 1",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /tree:
    post:
      summary: Create tree node
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TreeNode'
      responses:
        '201':
          description: Created
components:
  schemas:
    TreeNode:
      type: object
      properties:
        value:
          type: string
          description: Node value
        children:
          type: array
          items:
            $ref: '#/components/schemas/TreeNode'
          description: Child nodes`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"**value** (string)",
				"Node value",
				"**children** (array of objects)",
				"Child nodes",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertFieldDefinitionsArrayOfObjects(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "array of objects with nested schema",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /items:
    post:
      summary: Create items
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ItemList'
      responses:
        '201':
          description: Created
components:
  schemas:
    ItemList:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/Item'
          description: List of items
    Item:
      type: object
      required:
        - id
        - name
      properties:
        id:
          type: string
          description: Item ID
        name:
          type: string
          description: Item name`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"**items** (array of objects)",
				"List of items",
				"**Item**",
				"- `id` (string, required): Item ID",
				"- `name` (string, required): Item name",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertFieldDefinitionsMaxDepth(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "maximum depth limit enforced",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /deep:
    post:
      summary: Create deep structure
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/L1'
      responses:
        '201':
          description: Created
components:
  schemas:
    L1:
      type: object
      properties:
        l2:
          $ref: '#/components/schemas/L2'
    L2:
      type: object
      properties:
        l3:
          $ref: '#/components/schemas/L3'
    L3:
      type: object
      properties:
        l4:
          $ref: '#/components/schemas/L4'
    L4:
      type: object
      properties:
        l5:
          $ref: '#/components/schemas/L5'
    L5:
      type: object
      properties:
        l6:
          $ref: '#/components/schemas/L6'
    L6:
      type: object
      properties:
        l7:
          $ref: '#/components/schemas/L7'
    L7:
      type: object
      properties:
        l8:
          $ref: '#/components/schemas/L8'
    L8:
      type: object
      properties:
        l9:
          $ref: '#/components/schemas/L9'
    L9:
      type: object
      properties:
        l10:
          $ref: '#/components/schemas/L10'
    L10:
      type: object
      properties:
        value:
          type: string
          description: Deep value`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"**L2**",
				"**L3**",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertResponseFieldDefinitions2xx(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "200 response with field definitions",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      required:
        - id
      properties:
        id:
          type: string
          description: User identifier
        name:
          type: string
          description: User name`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Responses",
				"#### 200 Response",
				"Success",
				"```json",
				"#### Field Definitions",
				"**id** (string, required)",
				"User identifier",
				"**name** (string)",
				"User name",
			},
		},
		{
			name: "201 response with field definitions",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUser'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    CreateUser:
      type: object
      properties:
        name:
          type: string
    User:
      type: object
      properties:
        id:
          type: string
          description: User ID`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Responses",
				"#### 201 Response",
				"Created",
				"#### Field Definitions",
				"**id** (string)",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertResponseFieldDefinitions4xx(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
		notMd   []string
	}{
		{
			name: "400 response without field definitions",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      responses:
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    Error:
      type: object
      properties:
        code:
          type: string
          description: Error code
        message:
          type: string
          description: Error message`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Responses",
				"#### 400 Response",
				"Bad request",
				"```json",
			},
			notMd: []string{
				"#### Field Definitions",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			for _, notWant := range test.notMd {
				assert.NotContains(t, md, notWant)
			}
		})
	}
}

func TestConvertResponsesSectionHeader(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "responses section header present",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Success`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"### Responses",
				"#### 200 Response",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertResponseHeadingLevel(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
		notMd   []string
	}{
		{
			name: "response uses H4 not H5",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      responses:
        '200':
          description: Success
        '404':
          description: Not found`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### 200 Response",
				"#### 404 Response",
			},
			notMd: []string{
				"##### 200 Response",
				"##### 404 Response",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			for _, notWant := range test.notMd {
				assert.NotContains(t, md, notWant)
			}
		})
	}
}

func TestConvertResponseSharedSchema(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "shared schema in multiple 2xx responses",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      responses:
        '200':
          description: Success (existing user)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '201':
          description: Created (new user)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
          description: User ID
        name:
          type: string
          description: User name`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### 200 Response",
				"Success (existing user)",
				"#### 201 Response",
				"Created (new user)",
				"#### Field Definitions (applies to 200, 201 responses)",
				"**id** (string)",
				"**name** (string)",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertPathParametersFieldDef(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "path parameter with field definitions format",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      summary: Get user
      parameters:
        - name: id
          in: path
          required: true
          description: User identifier
          schema:
            type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Path Parameters",
				"`id` (string, required)",
				"User identifier",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertQueryParametersFieldDef(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "query parameter with field definitions format",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: limit
          in: query
          required: false
          description: Maximum number of items to return
          schema:
            type: integer`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Query Parameters",
				"**limit** (integer)",
				"Maximum number of items to return",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertHeadersTableFormat(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "headers remain in table format",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: X-API-Key
          in: header
          required: true
          description: API authentication key
          schema:
            type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Headers",
				"Name | Description | Required | Type",
				"-----|-------------|----------|-----",
				"X-API-Key | API authentication key | true | string",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertParameterWithEnums(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "query parameter with enum values",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      parameters:
        - name: status
          in: query
          required: false
          description: Filter by user status
          schema:
            type: string
            enum:
              - active
              - inactive
              - pending`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Query Parameters",
				"**status** (string)",
				"Filter by user status Enums: `active`, `inactive`, `pending`",
			},
		},
		{
			name: "path parameter with enum values",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /resource/{type}:
    get:
      summary: Get resource
      parameters:
        - name: type
          in: path
          required: true
          description: Type of resource
          schema:
            type: string
            enum:
              - user
              - group
              - role`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Path Parameters",
				"`type` (string, required)",
				"Type of resource Enums: `user`, `group`, `role`",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertParameterRequired(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "required vs optional parameters",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      summary: Get user
      parameters:
        - name: id
          in: path
          required: true
          description: User ID
          schema:
            type: string
        - name: expand
          in: query
          required: false
          description: Expand related objects
          schema:
            type: boolean`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"`id` (string, required)",
				"**expand** (boolean)",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertSharedSchemaAcrossEndpoints(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "schema used in multiple endpoints",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
  /users/{id}:
    get:
      summary: Get user
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: Not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    put:
      summary: Update user
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '200':
          description: Updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: Not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
      required:
        - id
    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"## Shared Schema Definitions",
				"### User",
				"### Error",
				"Used in: GET /users/{id}, POST /users, PUT /users/{id}",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertFieldDefinitionsReferenceShared(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "endpoint references shared schema",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
  /users/{id}:
    get:
      summary: Get user
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        name:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"#### Field Definitions",
				"See [User](#user)",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}
		})
	}
}

func TestConvertNonSharedSchemaInline(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantMd  []string
	}{
		{
			name: "schema used in single endpoint only",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateUserRequest'
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string
          description: User's full name
        email:
          type: string
          description: Email address
      required:
        - name
        - email`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: []string{
				"**name** (string, required)",
				"User's full name",
				"**email** (string, required)",
				"Email address",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, want := range test.wantMd {
				assert.Contains(t, md, want)
			}

			assert.NotContains(t, md, "## Shared Schema Definitions")
		})
	}
}

func TestConvertSharedDefinitionsSectionPlacement(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
	}{
		{
			name: "shared definitions appear after TOC",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
  /users/{id}:
    get:
      summary: Get user
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			tocIdx := strings.Index(md, "## Table of Contents")
			sharedIdx := strings.Index(md, "## Shared Schema Definitions")
			endpointIdx := strings.Index(md, "## POST /users")

			require.NotEqual(t, -1, tocIdx, "TOC should be present")
			require.NotEqual(t, -1, sharedIdx, "Shared definitions should be present")
			require.NotEqual(t, -1, endpointIdx, "Endpoint should be present")

			assert.True(t, tocIdx < endpointIdx, "TOC should come before endpoints")
			assert.True(t, endpointIdx < sharedIdx, "Endpoints should come before shared definitions")
		})
	}
}

func TestConvertSharedSchemaWithinEndpoint(t *testing.T) {
	for _, test := range []struct {
		name      string
		openapi   string
		opts      conv.ConvertOptions
		notWantMd []string
	}{
		{
			name: "schema used in request and response of same endpoint not shared",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateUserRequest'
components:
  schemas:
    CreateUserRequest:
      type: object
      properties:
        name:
          type: string`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			notWantMd: []string{
				"## Shared Schema Definitions",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.NoError(t, err)
			md := string(result.Markdown)

			for _, notWant := range test.notWantMd {
				assert.NotContains(t, md, notWant)
			}
		})
	}
}
