// Copyright 2017 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package properties

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Encoding specifies encoding of the input data.
type Encoding uint

const (
	// UTF8 interprets the input data as UTF-8.
	UTF8 Encoding = 1 << iota

	// ISO_8859_1 interprets the input data as ISO-8859-1.
	ISO_8859_1
)

type Loader struct {
	// Encoding determines how the data from files and byte buffers
	// is interpreted. For URLs the Content-Type header is used
	// to determine the encoding of the data.
	Encoding Encoding

	// DisableExpansion configures the property expansion of the
	// returned property object. When set to true, the property values
	// will not be expanded and the Property object will not be checked
	// for invalid expansion expressions.
	DisableExpansion bool

	// IgnoreMissing configures whether missing files or URLs which return
	// 404 are reported as errors. When set to true, missing files and 404
	// status codes are not reported as errors.
	IgnoreMissing bool
}

// Load reads a buffer into a Properties struct.
func (l *Loader) LoadBytes(buf []byte) (*Properties, error) {
	p, err := parse(convert(buf, l.Encoding))
	if err != nil {
		return nil, err
	}
	if l.DisableExpansion {
		return p, nil
	}
	return p, p.check()
}

// LoadAll reads the content of multiple URLs or files in the given order into
// a Properties struct. If IgnoreMissing is true then a 404 status code or
// missing file will not be reported as error. Encoding sets the encoding for
// files. For the URLs see LoadURL for the Content-Type header and the
// encoding.
func (l *Loader) LoadAll(names []string) (*Properties, error) {
	result := NewProperties()
	for _, name := range names {
		n, err := expandName(name)
		if err != nil {
			return nil, err
		}
		var p *Properties
		if strings.HasPrefix(n, "http://") || strings.HasPrefix(n, "https://") {
			p, err = l.LoadURL(n)
		} else {
			p, err = l.LoadFile(n)
		}
		if err != nil {
			return nil, err
		}
		result.Merge(p)
	}
	if l.DisableExpansion {
		return result, nil
	}
	return result, result.check()
}

// LoadFile reads a file into a Properties struct.
// If IgnoreMissing is true then a missing file will not be
// reported as error.
func (l *Loader) LoadFile(filename string) (*Properties, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if l.IgnoreMissing && os.IsNotExist(err) {
			LogPrintf("properties: %s not found. skipping", filename)
			return NewProperties(), nil
		}
		return nil, err
	}
	p, err := parse(convert(data, l.Encoding))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// LoadURL reads the content of the URL into a Properties struct.
//
// The encoding is determined via the Content-Type header which
// should be set to 'text/plain'. If the 'charset' parameter is
// missing, 'iso-8859-1' or 'latin1' the encoding is set to
// ISO-8859-1. If the 'charset' parameter is set to 'utf-8' the
// encoding is set to UTF-8. A missing content type header is
// interpreted as 'text/plain; charset=utf-8'.
func (l *Loader) LoadURL(url string) (*Properties, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("properties: error fetching %q. %s", url, err)
	}
	if resp.StatusCode == 404 && l.IgnoreMissing {
		LogPrintf("properties: %s returned %d. skipping", url, resp.StatusCode)
		return NewProperties(), nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("properties: %s returned %d", url, resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("properties: %s error reading response. %s", url, err)
	}
	if err = resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("properties: %s error reading response. %s", url, err)
	}

	ct := resp.Header.Get("Content-Type")
	var enc Encoding
	switch strings.ToLower(ct) {
	case "text/plain", "text/plain; charset=iso-8859-1", "text/plain; charset=latin1":
		enc = ISO_8859_1
	case "", "text/plain; charset=utf-8":
		enc = UTF8
	default:
		return nil, fmt.Errorf("properties: invalid content type %s", ct)
	}

	p, err := parse(convert(body, enc))
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Load reads a buffer into a Properties struct.
func Load(buf []byte, enc Encoding) (*Properties, error) {
	return loadBuf(buf, enc, false)
}

// LoadString reads an UTF8 string into a properties struct.
func LoadString(s string) (*Properties, error) {
	return loadBuf([]byte(s), UTF8, false)
}

// LoadMap creates a new Properties struct from a string map.
func LoadMap(m map[string]string) *Properties {
	p := NewProperties()
	for k, v := range m {
		p.Set(k, v)
	}
	return p
}

// LoadFile reads a file into a Properties struct.
func LoadFile(filename string, enc Encoding) (*Properties, error) {
	return loadAll([]string{filename}, enc, false, false)
}

// LoadFiles reads multiple files in the given order into
// a Properties struct. If 'ignoreMissing' is true then
// non-existent files will not be reported as error.
func LoadFiles(filenames []string, enc Encoding, ignoreMissing bool) (*Properties, error) {
	return loadAll(filenames, enc, ignoreMissing, false)
}

// LoadURL reads the content of the URL into a Properties struct.
// See Loader#LoadURL for details.
func LoadURL(url string) (*Properties, error) {
	return loadAll([]string{url}, UTF8, false, false)
}

// LoadURLs reads the content of multiple URLs in the given order into a
// Properties struct. If IgnoreMissing is true then a 404 status code will
// not be reported as error. See Loader#LoadURL for the Content-Type header
// and the encoding.
func LoadURLs(urls []string, ignoreMissing bool) (*Properties, error) {
	return loadAll(urls, UTF8, ignoreMissing, false)
}

// LoadAll reads the content of multiple URLs or files in the given order into a
// Properties struct. If 'ignoreMissing' is true then a 404 status code or missing file will
// not be reported as error. Encoding sets the encoding for files. For the URLs please see
// LoadURL for the Content-Type header and the encoding.
func LoadAll(names []string, enc Encoding, ignoreMissing bool) (*Properties, error) {
	return loadAll(names, enc, ignoreMissing, false)
}

// MustLoadString reads an UTF8 string into a Properties struct and
// panics on error.
func MustLoadString(s string) *Properties {
	return must(LoadString(s))
}

// MustLoadFile reads a file into a Properties struct and
// panics on error.
func MustLoadFile(filename string, enc Encoding) *Properties {
	return must(LoadFile(filename, enc))
}

// MustLoadFiles reads multiple files in the given order into
// a Properties struct and panics on error. If 'ignoreMissing'
// is true then non-existent files will not be reported as error.
func MustLoadFiles(filenames []string, enc Encoding, ignoreMissing bool) *Properties {
	return must(LoadFiles(filenames, enc, ignoreMissing))
}

// MustLoadURL reads the content of a URL into a Properties struct and
// panics on error.
func MustLoadURL(url string) *Properties {
	return must(LoadURL(url))
}

// MustLoadURLs reads the content of multiple URLs in the given order into a
// Properties struct and panics on error. If 'ignoreMissing' is true then a 404
// status code will not be reported as error.
func MustLoadURLs(urls []string, ignoreMissing bool) *Properties {
	return must(LoadURLs(urls, ignoreMissing))
}

// MustLoadAll reads the content of multiple URLs or files in the given order into a
// Properties struct. If 'ignoreMissing' is true then a 404 status code or missing file will
// not be reported as error. Encoding sets the encoding for files. For the URLs please see
// LoadURL for the Content-Type header and the encoding. It panics on error.
func MustLoadAll(names []string, enc Encoding, ignoreMissing bool) *Properties {
	return must(LoadAll(names, enc, ignoreMissing))
}

func loadBuf(buf []byte, enc Encoding, disableExpansion bool) (*Properties, error) {
	l := &Loader{Encoding: enc, DisableExpansion: disableExpansion}
	return l.LoadBytes(buf)
}

func loadAll(names []string, enc Encoding, ignoreMissing, disableExpansion bool) (*Properties, error) {
	l := &Loader{Encoding: enc, DisableExpansion: disableExpansion, IgnoreMissing: ignoreMissing}
	return l.LoadAll(names)
}

func loadFile(filename string, enc Encoding, ignoreMissing bool) (*Properties, error) {
	l := &Loader{Encoding: enc, IgnoreMissing: ignoreMissing}
	return l.LoadFile(filename)
}

func loadURL(url string, ignoreMissing bool) (*Properties, error) {
	l := &Loader{IgnoreMissing: ignoreMissing}
	return l.LoadURL(url)
}

func must(p *Properties, err error) *Properties {
	if err != nil {
		ErrorHandler(err)
	}
	return p
}

// expandName expands ${ENV_VAR} expressions in a name.
// If the environment variable does not exist then it will be replaced
// with an empty string. Malformed expressions like "${ENV_VAR" will
// be reported as error.
func expandName(name string) (string, error) {
	return expand(name, []string{}, "${", "}", make(map[string]string))
}

// Interprets a byte buffer either as an ISO-8859-1 or UTF-8 encoded string.
// For ISO-8859-1 we can convert each byte straight into a rune since the
// first 256 unicode code points cover ISO-8859-1.
func convert(buf []byte, enc Encoding) string {
	switch enc {
	case UTF8:
		return string(buf)
	case ISO_8859_1:
		runes := make([]rune, len(buf))
		for i, b := range buf {
			runes[i] = rune(b)
		}
		return string(runes)
	default:
		ErrorHandler(fmt.Errorf("unsupported encoding %v", enc))
	}
	panic("ErrorHandler should exit")
}
