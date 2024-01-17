package index2_test

import (
	"strings"
	"testing"

	"github.com/google/codesearch/index2"
)

func TestFile(t *testing.T) {
	f := index2.File()
	if !strings.HasSuffix(f, "/.csearchindex2") {
		t.Fatalf("File() returned %q, should end with /.csearchindex2", f)
	}
}
