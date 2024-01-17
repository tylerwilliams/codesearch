package index2

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cockroachdb/pebble"
)

// File returns the name of the index file to use.
// It is either $CSEARCHINDEX or $HOME/.csearchindex.
func File() string {
        f := os.Getenv("CSEARCHINDEX2")
        if f != "" {
                return f
        }
        var home string
        home = os.Getenv("HOME")
        if runtime.GOOS == "windows" && home == "" {
                home = os.Getenv("USERPROFILE")
        }
        return filepath.Clean(home + "/.csearchindex2")
}

// An Index implements read-only access to a trigram index.
type Index struct {
	db *pebble.DB
}

// Open returns a new Index for reading.
func Open(pebbleDir string) *Index {
	db, err := pebble.Open(pebbleDir, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	printDB(db)
	return &Index{db}
}

func (i *Index) Close() {
	if err := i.db.Close(); err != nil {
		log.Fatal(err)
	}
}
