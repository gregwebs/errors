//go:build go1.23

package errors

import "testing"

func testFindErr(err error, v int) bool {
	for err := range UnwrapGroups(err) {
		e := err.(*errWalkTest)
		if e.v == v {
			return true
		}
	}
	return false
}

func TestUnwrapGroups(t *testing.T) {
	err := &errWalkTest{
		sub: []error{
			&errWalkTest{
				v:     10,
				cause: &errWalkTest{v: 11},
			},
			&errWalkTest{
				v:     20,
				cause: &errWalkTest{v: 21, cause: &errWalkTest{v: 22}},
			},
			&errWalkTest{
				v:     30,
				cause: &errWalkTest{v: 31},
			},
		},
	}

	if !testFindErr(err, 11) {
		t.Errorf("not found in first cause chain")
	}

	if !testFindErr(err, 22) {
		t.Errorf("not found in siblings")
	}

	if testFindErr(err, 32) {
		t.Errorf("found not exists")
	}
}
