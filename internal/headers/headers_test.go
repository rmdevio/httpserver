package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFoo: Bar\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "Bar", headers.Get("Foo"))
	assert.Equal(t, "", headers.Get("MissingKey"))
	assert.Equal(t, 33, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid headers with mixed-case keys
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nFoo-Bar: Baz\r\nX-CuStOm-HeAdEr: Value\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("host"))
	assert.Equal(t, "Baz", headers.Get("foo-bar"))
	assert.Equal(t, "Value", headers.Get("x-custom-header"))
	assert.Equal(t, "", headers.Get("missingkey"))
	assert.Equal(t, 61, n)
	assert.False(t, done)

	// Test: Invalid spacing in header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple values for the same header
	headers = NewHeaders()
	data = []byte("Set-Header: header1\r\nSet-Header: header2\r\nSet-Header: header3\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "header1,header2,header3", headers.Get("Set-Header"))
	assert.Equal(t, 63, n)
	assert.False(t, done)

	// Test: No data
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, headers.headers, map[string]string{})
	assert.Equal(t, 2, n)
	assert.True(t, done)
}
