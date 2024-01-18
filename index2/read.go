package index2

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/RoaringBitmap/roaring"
	"github.com/cockroachdb/pebble"
	"github.com/google/codesearch/query"
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
	db      *pebble.DB
	Verbose bool
}

// Open returns a new Index for reading.
func Open(pebbleDir string) *Index {
	db, err := pebble.Open(pebbleDir, &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	//printDB(db)
	return &Index{
		db: db,
	}
}

func (i *Index) Close() {
	if err := i.db.Close(); err != nil {
		log.Fatal(err)
	}
}

// Name returns the name corresponding to the given fileid.
func (ix *Index) Name(fileid uint32) string {
	return string(ix.NameBytes(fileid))
}

// NameBytes returns the name corresponding to the given fileid.
func (ix *Index) NameBytes(fileid uint32) []byte {
	iter := ix.db.NewIter(&pebble.IterOptions{
		LowerBound: filenameKey(""),
		UpperBound: filenameKey(string('\xff')),
	})
	defer iter.Close()

	filePrefix := filenameKey(fmt.Sprintf("%x", string(uint32ToBytes(fileid))))
	if !iter.SeekGE(filePrefix) || !bytes.HasPrefix(iter.Key(), filePrefix) {
		log.Fatalf("File %d not found in index (prefix: %q)", fileid, filePrefix)
		return nil
	}
	buf := make([]byte, len(iter.Value()))
	copy(buf, iter.Value())
	return buf
}

// Paths returns the list of indexed paths.
func (ix *Index) Paths() []string {
	fileIDs := ix.allIndexedFiles()
	names := make([]string, 0, len(fileIDs))

	for _, fileID := range fileIDs {
		names = append(names, ix.Name(fileID))
	}
	return names
}

func (ix *Index) allIndexedFiles() []uint32 {
	iter := ix.db.NewIter(&pebble.IterOptions{
		LowerBound: filenameKey(""),
		UpperBound: filenameKey(string('\xff')),
	})
	defer iter.Close()

	found := make([]uint32, 0)
	for iter.First(); iter.Valid(); iter.Next() {
		digest := bytes.TrimPrefix(iter.Key(), filenameKey(""))
		hashSum, err := hex.DecodeString(string(digest))
		if err != nil {
			log.Fatal(err)
		}
		fileid := bytesToUint32(hashSum[:4])
		found = append(found, fileid)
	}
	return found
}

func (ix *Index) PostingList(trigram uint32) []uint32 {
	return ix.postingList(trigram, nil)
}

func (ix *Index) postingListBM(trigram uint32, restrict *roaring.Bitmap) *roaring.Bitmap {
	triString := trigramToString(trigram)
	iter := ix.db.NewIter(&pebble.IterOptions{
		LowerBound: trigramKey(triString),
		UpperBound: trigramKey(triString + string('\xff')),
	})
	defer iter.Close()

	resultSet := roaring.New()
	postingList := roaring.New()
	for iter.First(); iter.Valid(); iter.Next() {
		//log.Printf("query %q matched key %q", triString, iter.Key())
		if _, err := postingList.ReadFrom(bytes.NewReader(iter.Value())); err != nil {
			log.Fatal(err)
		}
		resultSet = roaring.Or(resultSet, postingList)
		postingList.Clear()
	}
	resultSet.AndNot(restrict)
	return resultSet
}

func (ix *Index) postingList(trigram uint32, restrict []uint32) []uint32 {
	bm := ix.postingListBM(trigram, roaring.BitmapOf(restrict...))
	return bm.ToArray()
}

func (ix *Index) PostingAnd(list []uint32, trigram uint32) []uint32 {
	return ix.postingAnd(list, trigram, nil)
}

func (ix *Index) postingAnd(list []uint32, trigram uint32, restrict []uint32) []uint32 {
	bm := ix.postingListBM(trigram, roaring.BitmapOf(restrict...))
	bm.And(roaring.BitmapOf(list...))
	return bm.ToArray()
}

func (ix *Index) PostingOr(list []uint32, trigram uint32) []uint32 {
	return ix.postingOr(list, trigram, nil)
}

func (ix *Index) postingOr(list []uint32, trigram uint32, restrict []uint32) []uint32 {
	bm := ix.postingListBM(trigram, roaring.BitmapOf(restrict...))
	bm.Or(roaring.BitmapOf(list...))
	return bm.ToArray()
}

func (ix *Index) PostingQuery(q *query.Query) []uint32 {
	return ix.postingQuery(q, nil)
}

func (ix *Index) postingQuery(q *query.Query, restrict []uint32) (ret []uint32) {
	var list []uint32
	switch q.Op {
	case query.QNone:
		// nothing
	case query.QAll:
		if restrict != nil {
			return restrict
		}
		list = ix.allIndexedFiles()
	case query.QAnd:
		for _, t := range q.Trigram {
			tri := uint32(t[0])<<16 | uint32(t[1])<<8 | uint32(t[2])
			if list == nil {
				list = ix.postingList(tri, restrict)
			} else {
				list = ix.postingAnd(list, tri, restrict)
			}
			if len(list) == 0 {
				return nil
			}
		}
		for _, sub := range q.Sub {
			if list == nil {
				list = restrict
			}
			list = ix.postingQuery(sub, list)
			if len(list) == 0 {
				return nil
			}
		}
	case query.QOr:
		for _, t := range q.Trigram {
			tri := uint32(t[0])<<16 | uint32(t[1])<<8 | uint32(t[2])
			if list == nil {
				list = ix.postingList(tri, restrict)
			} else {
				list = ix.postingOr(list, tri, restrict)
			}
		}
		for _, sub := range q.Sub {
			list1 := ix.postingQuery(sub, restrict)
			list = mergeOr(list, list1)
		}
	}
	return list
}

func mergeOr(l1, l2 []uint32) []uint32 {
	var l []uint32
	i := 0
	j := 0
	for i < len(l1) || j < len(l2) {
		switch {
		case j == len(l2) || (i < len(l1) && l1[i] < l2[j]):
			l = append(l, l1[i])
			i++
		case i == len(l1) || (j < len(l2) && l1[i] > l2[j]):
			l = append(l, l2[j])
			j++
		case l1[i] == l2[j]:
			l = append(l, l1[i])
			i++
			j++
		}
	}
	return l
}
