// Copyright 2013 Frank Schroeder. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package goproperties

import (
	"fmt"
	"io/ioutil"
	"os"

	. "launchpad.net/gocheck"
)

type LoadSuite struct {
	tempFiles []string
}

var (
	_ = Suite(&LoadSuite{})
)

// ----------------------------------------------------------------------------

func (s *LoadSuite) TestLoadFailsWithNotExistingFile(c *C) {
	_, err := LoadFile("doesnotexist.properties")
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "open.*no such file or directory")
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) TestLoadFilesFailsOnNotExistingFile(c *C) {
	_, err := LoadFiles([]string{"doesnotexist.properties"}, false)
	c.Assert(err, NotNil)
	c.Assert(err, ErrorMatches, "open.*no such file or directory")
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) TestLoadFilesDoesNotFailOnNotExistingFileAndIgnoreMissing(c *C) {
	p, err := LoadFiles([]string{"doesnotexist.properties"}, true)
	c.Assert(err, IsNil)
	c.Assert(p.Len(), Equals, 0)
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) TestLoad(c *C) {
	filename := s.makeFile(c, "key=value")
	p := MustLoadFile(filename)

	c.Assert(p.Len(), Equals, 1)
	assertKeyValues(c, "", p, "key", "value")
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) TestLoadFiles(c *C) {
	filename := s.makeFile(c, "key=value")
	filename2 := s.makeFile(c, "key2=value2")
	p := MustLoadFiles([]string{filename, filename2}, false)
	assertKeyValues(c, "", p, "key", "value", "key2", "value2")
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) TestLoadFilesAndIgnoreMissing(c *C) {
	filename := s.makeFile(c, "key=value")
	filename2 := s.makeFile(c, "key2=value2")
	p := MustLoadFiles([]string{filename, filename + "foo", filename2, filename2 + "foo"}, true)
	assertKeyValues(c, "", p, "key", "value", "key2", "value2")
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) SetUpSuite(c *C) {
	s.tempFiles = make([]string, 0)
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) TearDownSuite(c *C) {
	for _, path := range s.tempFiles {
		err := os.Remove(path)
		if err != nil {
			fmt.Printf("os.Remove: %v", err)
		}
	}
}

// ----------------------------------------------------------------------------

func (s *LoadSuite) makeFile(c *C, data string) string {
	f, err := ioutil.TempFile("", "properties")
	if err != nil {
		fmt.Printf("ioutil.TempFile: %v", err)
		c.FailNow()
	}

	// remember the temp file so that we can remove it later
	s.tempFiles = append(s.tempFiles, f.Name())

	n, err := fmt.Fprint(f, data)
	if err != nil {
		fmt.Printf("fmt.Fprintln: %v", err)
		c.FailNow()
	}
	if n != len(data) {
		fmt.Printf("Data size mismatch. expected=%d wrote=%d\n", len(data), n)
		c.FailNow()
	}

	err = f.Close()
	if err != nil {
		fmt.Printf("f.Close: %v", err)
		c.FailNow()
	}

	return f.Name()
}
