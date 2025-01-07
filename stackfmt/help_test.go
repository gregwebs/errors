package stackfmt

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func testFormatRegexp(t *testing.T, n int, arg interface{}, format, wantAll string) {
	t.Helper()
	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(wantAll, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("test %d: wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", n+1, len(wantLines), len(gotLines), got, wantLines)
		return
	}

	for i, wantLine := range wantLines {
		want := wantLine
		got := gotLines[i]
		adjustedGot := regexp.MustCompile(`\S.*/errors`).ReplaceAllString(got, `github.com/gregwebs/errors`)
		match, err := regexp.MatchString(want, adjustedGot)
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("test %d: line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", n+1, i+1, format, adjustedGot, want)
		}
	}
}

func testFormatString(t *testing.T, n int, arg interface{}, format, wantAll string) {
	t.Helper()
	got := fmt.Sprintf(format, arg)
	gotLines := strings.SplitN(got, "\n", -1)
	wantLines := strings.SplitN(wantAll, "\n", -1)

	if len(wantLines) > len(gotLines) {
		t.Errorf("test %d: wantLines(%d) > gotLines(%d):\n got: %q\nwant: %q", n+1, len(wantLines), len(gotLines), got, wantLines)
		return
	}

	for i, wantLine := range wantLines {
		want := wantLine
		got := gotLines[i]
		adjustedGot := regexp.MustCompile(`\S.*/errors`).ReplaceAllString(got, `github.com/gregwebs/errors`)
		if want != adjustedGot {
			t.Errorf("test %d: line %d: fmt.Sprintf(%q, err):\n got: %q\nwant: %q", n+1, i+1, format, adjustedGot, want)
		}
	}
}
