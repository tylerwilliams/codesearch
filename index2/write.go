package index2

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"github.com/google/codesearch/sparse"
	"github.com/cockroachdb/pebble"
	"github.com/RoaringBitmap/roaring"
)

const (
	dataPrefix     = "dat:"
	filenamePrefix = "fil:"
	trigramPrefix  = "tri:"
)

type IndexWriter struct {
	db *pebble.DB

	LogSkip bool // log information about skipped files
	Verbose bool // log status using package log
	
	inbuf []byte        // input buffer
	totalBytes int64

	trigram *sparse.Set // trigrams for the current file
	postingLists map[uint32]*roaring.Bitmap
}

func printDB(db *pebble.DB) {
       iter := db.NewIter(&pebble.IterOptions{
               LowerBound: []byte{0},
               UpperBound: []byte{math.MaxUint8},
       })
       defer iter.Close()
       log.Printf("<BEGIN DB>")
       for iter.First(); iter.Valid(); iter.Next() {
               log.Printf("\tkey: %q len(val): %d\n", string(iter.Key()), len(iter.Value()))
       }
       log.Printf("<END DB>")
}

// validUTF8 reports whether the byte pair can appear in a
// valid sequence of UTF-8-encoded code points.
func validUTF8(c1, c2 uint32) bool {
	switch {
	case c1 < 0x80:
		// 1-byte, must be followed by 1-byte or first of multi-byte
		return c2 < 0x80 || 0xc0 <= c2 && c2 < 0xf8
	case c1 < 0xc0:
		// continuation byte, can be followed by nearly anything
		return c2 < 0xf8
	case c1 < 0xf8:
		// first of multi-byte, must be followed by continuation byte
		return 0x80 <= c2 && c2 < 0xc0
	}
	return false
}

func uint32ToBytes(i uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, i)
	return buf
}

func bytesToUint32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

// Tuning constants for detecting text files.
// A file is assumed not to be text files (and thus not indexed)
// if it contains an invalid UTF-8 sequences, if it is longer than maxFileLength
// bytes, if it contains a line longer than maxLineLen bytes,
// or if it contains more than maxTextTrigrams distinct trigrams.
const (
	maxFileLen      = 1 << 30
	maxLineLen      = 2000
	maxTextTrigrams = 20000
)

 // Create returns a new IndexWriter that will write the index to file.
func Create(pebbleDir string) *IndexWriter {
	log.Printf("Opening pebbleDir %q", pebbleDir)
	db, err := pebble.Open(pebbleDir, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	printDB(db)
	return &IndexWriter{
		db: db,
		trigram:   sparse.NewSet(1 << 24),
		postingLists: make(map[uint32]*roaring.Bitmap),
		inbuf:     make([]byte, 16384),
	}
}	

// AddPaths adds the given paths to the index's list of paths.
func (iw *IndexWriter) AddPaths(paths []string) {
	log.Printf("AddPaths %s (ignored?!?!)", paths)
	return
}

// AddFile adds the file with the given name (opened using os.Open)
// to the index. It logs errors using package log.
func (iw *IndexWriter) AddFile(name string) {
	f, err := os.Open(name)
	if err != nil {
		log.Print(err)
		return
	}
	defer f.Close()
	iw.Add(name, f)
}

// Add adds the file f to the index under the given name.
// It logs errors using package log.
func (iw *IndexWriter) Add(name string, f io.ReadSeeker) {
	// Compute the SHA256 hash of the file.
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}
	hashSum := h.Sum(nil)
	digest := fmt.Sprintf("%x", hashSum)

	fileid := bytesToUint32(hashSum[:4])
	filenameKey := []byte(filenamePrefix + digest)
	_, closer, err := iw.db.Get(filenameKey)
	if err != pebble.ErrNotFound {
		log.Printf("File %q already indexed!!!", name)
		closer.Close()
		return
	}
	log.Printf("Indexing file %q hash: %s id: %d", name, filenameKey, fileid)
	f.Seek(0, 0)
	
	iw.trigram.Reset()
	var (
		c       = byte(0)
		i       = 0
		buf     = iw.inbuf[:0]
		tv      = uint32(0)
		n       = int64(0)
		linelen = 0
	)
	for {
		tv = (tv << 8) & (1<<24 - 1)
		if i >= len(buf) {
			n, err := f.Read(buf[:cap(buf)])
			if n == 0 {
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Printf("%s: %v\n", name, err)
					return
				}
				log.Printf("%s: 0-length read\n", name)
				return
			}
			buf = buf[:n]
			i = 0
		}
		c = buf[i]
		i++
		tv |= uint32(c)
		if n++; n >= 3 {
			iw.trigram.Add(tv)
		}
		if !validUTF8((tv>>8)&0xFF, tv&0xFF) {
			if iw.LogSkip {
				log.Printf("%s: invalid UTF-8, ignoring\n", name)
			}
			return
		}
		if n > maxFileLen {
			if iw.LogSkip {
				log.Printf("%s: too long, ignoring\n", name)
			}
			return
		}
		if linelen++; linelen > maxLineLen {
			if iw.LogSkip {
				log.Printf("%s: very long lines, ignoring\n", name)
			}
			return
		}
		if c == '\n' {
			linelen = 0
		}
	}
	if iw.trigram.Len() > maxTextTrigrams {
		if iw.LogSkip {
			log.Printf("%s: too many trigrams, probably not text, ignoring\n", name)
		}
		return
	}
	iw.totalBytes += n

	if iw.Verbose {
		log.Printf("%d %d %s\n", n, iw.trigram.Len(), name)
	}

	for _, trigram := range iw.trigram.Dense() {
		pl, ok := iw.postingLists[trigram]
		if !ok {
			pl = roaring.New()
			iw.postingLists[trigram] = pl
		}
		pl.Add(fileid)
	}
	if err := iw.db.Set(filenameKey, []byte(name), pebble.NoSync); err != nil {
		log.Fatal(err)
	}
	dataKey := []byte(dataPrefix + digest)
	if err := iw.db.Set(dataKey, []byte(buf), pebble.NoSync); err != nil {
		log.Fatal(err)
	}
}

func (iw *IndexWriter) Close() {
	if err := iw.db.Close(); err != nil {
		log.Fatal(err)
	}
}

func trigramToBytes(tv uint32) []byte {
       l := byte((tv >> 16) & 255)
       m := byte((tv >> 8) & 255)
       r := byte(tv & 255)
       return []byte{l,m,r}
}

func (iw *IndexWriter) Flush() {
	for trigram, pl := range iw.postingLists {
		pl.RunOptimize()
		trigramKey := append([]byte(trigramPrefix), trigramToBytes(trigram)...)
		buf := new(bytes.Buffer)
		if _, err := pl.WriteTo(buf); err != nil {
			log.Fatal(err)
		}
		if err := iw.db.Set(trigramKey, buf.Bytes(), pebble.NoSync); err != nil {
			log.Fatal(err)
		}
	}
	if err := iw.db.Flush(); err != nil {
		log.Fatal(err)
	}
}

