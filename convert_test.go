package conv_test

import (
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
			wantMd: "# Test API\n\nTest Description\n",
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
