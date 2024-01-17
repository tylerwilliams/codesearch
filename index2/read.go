package index2

import (
	"bytes"
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
		LowerBound: []byte(filenamePrefix),
		UpperBound: []byte(filenamePrefix + string('\xff')),
	})
	defer iter.Close()

	filePrefix := []byte(fmt.Sprintf("%s%x", filenamePrefix, string(uint32ToBytes(fileid))))
	if !iter.SeekGE(filePrefix) || !bytes.HasPrefix(iter.Key(), filePrefix) {
		log.Fatalf("File %d not found in index (prefix: %q)", fileid, filePrefix)
		return nil
	}
	buf := make([]byte, len(iter.Value()))
	copy(buf, iter.Value())
	return buf
}

func (ix *Index) PostingList(trigram uint32) []uint32 {
	return ix.postingList(trigram, nil)
}

func (ix *Index) postingListBM(trigram uint32, restrict *roaring.Bitmap) *roaring.Bitmap {
	trigramKey := append([]byte(trigramPrefix), trigramToBytes(trigram)...)
	buf, closer, err := ix.db.Get(trigramKey)
	if err == pebble.ErrNotFound {
		return nil
	} else if err != nil {
		log.Fatal(err)
	}
	defer closer.Close()
	bm := roaring.New()
	if _, err := bm.ReadFrom(bytes.NewReader(buf)); err != nil {
		log.Fatal(err)
	}
	bm.AndNot(restrict)
	return bm
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
		log.Fatalf("QAll NOT SUPPORTED")
		//		list = make([]uint32, ix.numName)
		//		for i := range list {
		//			list[i] = uint32(i)
		//		}
		//		return list
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

func corrupt() {
	log.Fatal("corrupt index: remove " + File())
}
