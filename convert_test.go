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

func TestConvertDebugParameterExtraction(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
        - name: X-API-Key
          in: header
          required: true
          schema:
            type: string
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
		Debug: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Debug)

	assert.Equal(t, 1, result.Debug.ParameterCounts["path"])
	assert.Equal(t, 1, result.Debug.ParameterCounts["query"])
	assert.Equal(t, 1, result.Debug.ParameterCounts["header"])
}

func TestConvertDebugResponseExtraction(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: Success
        '400':
          description: Bad request
    post:
      responses:
        '201':
          description: Created
        '400':
          description: Bad request
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
		Debug: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Debug)

	assert.Equal(t, 1, result.Debug.ResponseCounts["200"])
	assert.Equal(t, 1, result.Debug.ResponseCounts["201"])
	assert.Equal(t, 2, result.Debug.ResponseCounts["400"])
}

func TestConvertDebugTagGrouping(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      tags:
        - users
  /posts:
    get:
      tags:
        - posts
  /health:
    get:
      summary: Health check
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
		Debug: true,
	})
	require.NoError(t, err)
	require.NotNil(t, result.Debug)

	assert.Equal(t, 3, result.Debug.ParsedPaths)
	assert.Equal(t, 3, result.Debug.ExtractedOps)
	assert.Equal(t, 1, result.Debug.UntaggedOps)
	assert.Contains(t, result.Debug.TagsFound, "users")
	assert.Contains(t, result.Debug.TagsFound, "posts")
	assert.Contains(t, result.Debug.TagsFound, "Default APIs")
}

func TestConvertIntegrationPetStore(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Pet Store API
  description: A sample API for managing pets
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List all pets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: Maximum number of pets to return
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
  /pets/{petId}:
    get:
      summary: Get a pet by ID
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
        '404':
          description: Pet not found
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Pet Store API",
		Debug: true,
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "# Pet Store API")
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "## pets")

	assert.Equal(t, 2, result.EndpointCount)
	assert.Equal(t, 1, result.TagCount)

	require.NotNil(t, result.Debug)
	assert.Equal(t, 2, result.Debug.ParsedPaths)
	assert.Equal(t, 2, result.Debug.ExtractedOps)
	assert.Equal(t, 0, result.Debug.UntaggedOps)
	assert.Equal(t, 1, result.Debug.ParameterCounts["path"])
	assert.Equal(t, 1, result.Debug.ParameterCounts["query"])
}

func TestConvertIntegrationComplexAPI(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Complex API
  description: A comprehensive API with multiple features
  version: 2.0.0
paths:
  /users:
    get:
      summary: List all users
      tags:
        - users
      parameters:
        - name: limit
          in: query
          description: Maximum number of users
          required: false
          schema:
            type: integer
        - name: offset
          in: query
          description: Number of users to skip
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
        '400':
          description: Invalid parameters
    post:
      summary: Create a new user
      tags:
        - users
      parameters:
        - name: X-Request-ID
          in: header
          description: Request identifier
          required: true
          schema:
            type: string
      responses:
        '201':
          description: User created
        '400':
          description: Invalid user data
  /users/{userId}:
    get:
      summary: Get user by ID
      tags:
        - users
      parameters:
        - name: userId
          in: path
          description: User identifier
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: User not found
    delete:
      summary: Delete user
      tags:
        - users
        - admin
      parameters:
        - name: userId
          in: path
          description: User identifier
          required: true
          schema:
            type: string
        - name: X-Admin-Token
          in: header
          description: Admin authorization
          required: true
          schema:
            type: string
      responses:
        '200':
          description: User deleted
        '403':
          description: Unauthorized
        '404':
          description: User not found
  /posts:
    get:
      summary: List all posts
      tags:
        - posts
      responses:
        '200':
          description: Successful response
  /health:
    get:
      summary: Health check
      responses:
        '200':
          description: Service is healthy
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title:       "Complex API",
		Description: "A comprehensive API with multiple features",
		Debug:       true,
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	assert.Contains(t, markdown, "# Complex API")
	assert.Contains(t, markdown, "A comprehensive API with multiple features")
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "## users")
	assert.Contains(t, markdown, "## admin")
	assert.Contains(t, markdown, "## posts")
	assert.Contains(t, markdown, "## Default APIs")

	assert.Contains(t, markdown, "### GET /users")
	assert.Contains(t, markdown, "### POST /users")
	assert.Contains(t, markdown, "### GET /users/{userId}")
	assert.Contains(t, markdown, "### DELETE /users/{userId}")
	assert.Contains(t, markdown, "### GET /posts")
	assert.Contains(t, markdown, "### GET /health")

	assert.Contains(t, markdown, "#### Path Parameters")
	assert.Contains(t, markdown, "#### Query Parameters")
	assert.Contains(t, markdown, "#### Header Parameters")

	assert.Equal(t, 6, result.EndpointCount)
	assert.Equal(t, 4, result.TagCount)

	require.NotNil(t, result.Debug)
	assert.Equal(t, 4, result.Debug.ParsedPaths)
	assert.Equal(t, 6, result.Debug.ExtractedOps)
	assert.Equal(t, 1, result.Debug.UntaggedOps)
	assert.Equal(t, 2, result.Debug.ParameterCounts["path"])
	assert.Equal(t, 2, result.Debug.ParameterCounts["query"])
	assert.Equal(t, 2, result.Debug.ParameterCounts["header"])
	assert.Equal(t, 5, result.Debug.ResponseCounts["200"])
	assert.Equal(t, 2, result.Debug.ResponseCounts["400"])
	assert.Equal(t, 1, result.Debug.ResponseCounts["201"])
	assert.Equal(t, 1, result.Debug.ResponseCounts["403"])
	assert.Equal(t, 2, result.Debug.ResponseCounts["404"])
}

func TestConvertCompleteExample(t *testing.T) {
	openapi, err := os.ReadFile("examples/openapi.yaml")
	require.NoError(t, err)

	expected, err := os.ReadFile("examples/example.md")
	require.NoError(t, err)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title:       "Pet Store API",
		Description: "A comprehensive API for managing a pet store with users, pets, and orders",
	})
	require.NoError(t, err)

	assert.Equal(t, string(expected), string(result.Markdown))

	assert.Equal(t, 13, result.EndpointCount)
	assert.Equal(t, 5, result.TagCount)
}

func TestConvert_ResponseExamplePriority(t *testing.T) {
	for _, test := range []struct {
		name     string
		spec     []byte
		wantJSON string
	}{
		{
			name: "ExplicitExample",
			spec: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
              example:
                id: "123"
                text: "explicit example"
components:
  schemas:
    Message:
      type: object
      properties:
        id:
          type: string
        text:
          type: string
`),
			wantJSON: `"id": "123"`,
		},
		{
			name: "NamedExamples",
			spec: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
              examples:
                example1:
                  value:
                    id: "456"
                    text: "named example"
components:
  schemas:
    Message:
      type: object
      properties:
        id:
          type: string
        text:
          type: string
`),
			wantJSON: `"id": "456"`,
		},
		{
			name: "GeneratedFromRef",
			spec: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
components:
  schemas:
    Message:
      type: object
      properties:
        text:
          type: string
`),
			wantJSON: `"text":`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert(test.spec, conv.ConvertOptions{Title: "Test API"})
			require.NoError(t, err)
			assert.Contains(t, string(result.Markdown), test.wantJSON)
		})
	}
}

func TestConvert_ResponseSchemaErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		spec    []byte
		wantErr string
	}{
		{
			name: "InlineSchemaNotAllowed",
			spec: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  text:
                    type: string
`),
			wantErr: "inline schemas not supported in responses",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := conv.Convert(test.spec, conv.ConvertOptions{Title: "Test API"})
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func TestConvert_ResponseContentTypes(t *testing.T) {
	for _, test := range []struct {
		name         string
		spec         []byte
		wantContains string
		wantMissing  string
	}{
		{
			name: "NoContentDescriptionOnly",
			spec: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
`),
			wantContains: "Success",
			wantMissing:  "```json",
		},
		{
			name: "MultipleMediaTypesOnlyJSON",
			spec: []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
            text/html:
              schema:
                type: string
components:
  schemas:
    Message:
      type: object
      properties:
        text:
          type: string
`),
			wantContains: "```json",
			wantMissing:  "```html",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert(test.spec, conv.ConvertOptions{Title: "Test API"})
			require.NoError(t, err)
			markdown := string(result.Markdown)
			if test.wantContains != "" {
				assert.Contains(t, markdown, test.wantContains)
			}
			if test.wantMissing != "" {
				assert.NotContains(t, markdown, test.wantMissing)
			}
		})
	}
}

func TestCompleteExampleWithGoldenFile(t *testing.T) {
	spec, err := os.ReadFile("examples/openapi.yaml")
	require.NoError(t, err)

	result, err := conv.Convert(spec, conv.ConvertOptions{
		Title:       "Pet Store API",
		Description: "A comprehensive API for managing a pet store with users, pets, and orders",
	})
	require.NoError(t, err)

	golden, err := os.ReadFile("testdata/golden/petstore-example.md")
	require.NoError(t, err)

	if !bytes.Equal(result.Markdown, golden) {
		actualPath := "testdata/golden/petstore-example.actual.md"
		err := os.WriteFile(actualPath, result.Markdown, 0644)
		require.NoError(t, err)

		t.Fatalf("Output doesn't match golden file\nActual written to: %s\nRun: diff %s testdata/golden/petstore-example.md",
			actualPath, actualPath)
	}
}
