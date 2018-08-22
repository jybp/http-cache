// Package disk provides an implementation of httpcache.Cache that uses the filesystem to store http responses.
package disk

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	defaultDir                  = "/tmp/httpcache"
	defaultFilePerm os.FileMode = 0666
	defaultPathPerm os.FileMode = 0777
)

// Cache is an implementation of httpcache.Cache.
type Cache struct {
	Dir      string
	PathPerm os.FileMode
	FilePerm os.FileMode
}

// Get returns a dumped response if the file corresponding to key exists.
func (c Cache) Get(key string) ([]byte, bool) {
	b, err := ioutil.ReadFile(filepath.Join(c.Dir, key))
	if err != nil {
		return nil, false
	}
	return b, true
}

// Set stores a dumped response to disk with filename key.
func (c Cache) Set(key string, response []byte) {
	if _, err := os.Stat(c.Dir); err != nil {
		if c.PathPerm == 0 {
			c.PathPerm = defaultPathPerm
		}
		if err := os.MkdirAll(c.Dir, c.PathPerm); err != nil {
			return
		}
	}
	if c.FilePerm == 0 {
		c.FilePerm = defaultFilePerm
	}
	if err := ioutil.WriteFile(filepath.Join(c.Dir, key), response, c.FilePerm); err != nil {
		return
	}
}
