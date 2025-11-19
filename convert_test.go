package conv_test

import (
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

func TestConvertInvalidOpenAPI(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantErr string
	}{
		{
			name:    "invalid yaml",
			openapi: "this is not valid yaml: {[}",
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantErr: "failed to parse openapi document",
		},
		{
			name: "invalid reference",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        $ref: "#/invalid/ref"`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantErr: "failed to build openapi 3.x model",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			require.ErrorContains(t, err, test.wantErr)
			require.Nil(t, result)
		})
	}
}

func TestConvertOpenAPI2Rejected(t *testing.T) {
	for _, test := range []struct {
		name    string
		openapi string
		opts    conv.ConvertOptions
		wantErr string
	}{
		{
			name: "openapi 2.0 spec",
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
	openapi := []byte(`openapi: 3.0.0
info:
  title: Pet Store
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List all pets
      tags:
        - pets
  /pets/{petId}:
    get:
      summary: Get a pet by ID
      tags:
        - pets
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Pet Store API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "# Pet Store API")
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "HTTP Request | Description")
	assert.Contains(t, markdown, "GET [/pets](#getpets) | List all pets")
	assert.Contains(t, markdown, "GET [/pets/{petId}](#getpetspetid) | Get a pet by ID")
	assert.Contains(t, markdown, "## pets")
	assert.Contains(t, markdown, "### GET /pets")
	assert.Contains(t, markdown, "List all pets")
	assert.Equal(t, 2, result.EndpointCount)
	assert.Equal(t, 1, result.TagCount)
}

func TestConvertSingleEndpoint(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users
      description: Returns a list of users
      tags:
        - users
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "# Test API")
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "GET [/users](#getusers) | Get users")
	assert.Contains(t, markdown, "## users")
	assert.Contains(t, markdown, "### GET /users")
	assert.Contains(t, markdown, "Get users")
	assert.Contains(t, markdown, "Returns a list of users")
	assert.Equal(t, 1, result.EndpointCount)
	assert.Equal(t, 1, result.TagCount)
}

func TestConvertMultipleEndpoints(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      tags:
        - users
    post:
      summary: Create user
      tags:
        - users
  /posts:
    get:
      summary: List posts
      tags:
        - posts
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "GET [/users](#getusers) | List users")
	assert.Contains(t, markdown, "POST [/users](#postusers) | Create user")
	assert.Contains(t, markdown, "GET [/posts](#getposts) | List posts")
	assert.Contains(t, markdown, "## posts")
	assert.Contains(t, markdown, "## users")
	assert.Contains(t, markdown, "### GET /users")
	assert.Contains(t, markdown, "### POST /users")
	assert.Contains(t, markdown, "### GET /posts")
	assert.Equal(t, 3, result.EndpointCount)
	assert.Equal(t, 2, result.TagCount)
}

func TestConvertUntaggedOperations(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /health:
    get:
      summary: Health check
  /metrics:
    get:
      summary: Get metrics
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "## Default APIs")
	assert.Contains(t, markdown, "### GET /health")
	assert.Contains(t, markdown, "Health check")
	assert.Contains(t, markdown, "### GET /metrics")
	assert.Contains(t, markdown, "Get metrics")
	assert.Equal(t, 2, result.EndpointCount)
	assert.Equal(t, 1, result.TagCount)
}

func TestConvertMultipleTagsPerOperation(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /admin/users:
    delete:
      summary: Delete user
      tags:
        - admin
        - users
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "## admin")
	assert.Contains(t, markdown, "## users")

	adminSection := markdown[strings.Index(markdown, "## admin"):]
	usersSection := markdown[strings.Index(markdown, "## users"):]

	assert.Contains(t, adminSection, "### DELETE /admin/users")
	assert.Contains(t, adminSection, "Delete user")
	assert.Contains(t, usersSection, "### DELETE /admin/users")
	assert.Contains(t, usersSection, "Delete user")
	assert.Equal(t, 1, result.EndpointCount)
	assert.Equal(t, 2, result.TagCount)
}

func TestConvertPathParameters(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets/{petId}:
    get:
      summary: Get a pet
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          description: ID of pet to return
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "#### Path Parameters")
	assert.Contains(t, markdown, "Name | Description | Required | Type")
	assert.Contains(t, markdown, "petId | ID of pet to return | true | string")
	assert.Contains(t, markdown, "##### 200 Response")
	assert.Contains(t, markdown, "Successful response")
}

func TestConvertQueryParameters(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List pets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: Maximum number of pets to return
          required: false
          schema:
            type: integer
        - name: offset
          in: query
          description: Number of pets to skip
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "#### Query Parameters")
	assert.Contains(t, markdown, "Name | Description | Required | Type")
	assert.Contains(t, markdown, "limit | Maximum number of pets to return | false | integer")
	assert.Contains(t, markdown, "offset | Number of pets to skip | false | integer")
}

func TestConvertHeaderParameters(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets:
    post:
      summary: Create a pet
      tags:
        - pets
      parameters:
        - name: X-Request-ID
          in: header
          description: Unique request identifier
          required: true
          schema:
            type: string
        - name: X-API-Key
          in: header
          description: API authentication key
          required: true
          schema:
            type: string
      responses:
        '201':
          description: Pet created successfully
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "#### Header Parameters")
	assert.Contains(t, markdown, "Name | Description | Required | Type")
	assert.Contains(t, markdown, "X-Request-ID | Unique request identifier | true | string")
	assert.Contains(t, markdown, "X-API-Key | API authentication key | true | string")
}

func TestConvertResponseExamples(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets/{petId}:
    get:
      summary: Get a pet
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: Pet not found
        '500':
          description: Internal server error
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "##### 200 Response")
	assert.Contains(t, markdown, "Successful response")
	assert.Contains(t, markdown, "##### 404 Response")
	assert.Contains(t, markdown, "Pet not found")
	assert.Contains(t, markdown, "##### 500 Response")
	assert.Contains(t, markdown, "Internal server error")
}

func TestConvertCompleteEndpoint(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{userId}/orders:
    get:
      summary: Get user orders
      description: Returns all orders placed by a specific user
      tags:
        - users
        - orders
      parameters:
        - name: userId
          in: path
          description: User identifier
          required: true
          schema:
            type: string
        - name: status
          in: query
          description: Filter by order status
          required: false
          schema:
            type: string
        - name: limit
          in: query
          description: Maximum number of orders to return
          required: false
          schema:
            type: integer
        - name: X-API-Key
          in: header
          description: API authentication key
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response with order list
        '400':
          description: Invalid request parameters
        '404':
          description: User not found
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "### GET /users/{userId}/orders")
	assert.Contains(t, markdown, "Get user orders")
	assert.Contains(t, markdown, "Returns all orders placed by a specific user")

	assert.Contains(t, markdown, "#### Path Parameters")
	assert.Contains(t, markdown, "userId | User identifier | true | string")

	assert.Contains(t, markdown, "#### Query Parameters")
	assert.Contains(t, markdown, "status | Filter by order status | false | string")
	assert.Contains(t, markdown, "limit | Maximum number of orders to return | false | integer")

	assert.Contains(t, markdown, "#### Header Parameters")
	assert.Contains(t, markdown, "X-API-Key | API authentication key | true | string")

	assert.Contains(t, markdown, "##### 200 Response")
	assert.Contains(t, markdown, "Successful response with order list")
	assert.Contains(t, markdown, "##### 400 Response")
	assert.Contains(t, markdown, "Invalid request parameters")
	assert.Contains(t, markdown, "##### 404 Response")
	assert.Contains(t, markdown, "User not found")

	assert.Contains(t, markdown, "## orders")
	assert.Contains(t, markdown, "## users")

	assert.Equal(t, 1, result.EndpointCount)
	assert.Equal(t, 2, result.TagCount)
}
