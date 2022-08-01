package tests

import (
	"anileha/util"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestFileIndicesCorrect(t *testing.T) {
	result, err := util.ParseFileIndices("0-4,8-9,6")
	if err != nil {
		t.Fatal(err)
	}
	expected := make(map[uint]struct{}, 10)
	expected[0] = struct{}{}
	expected[1] = struct{}{}
	expected[2] = struct{}{}
	expected[3] = struct{}{}
	expected[4] = struct{}{}
	expected[6] = struct{}{}
	expected[8] = struct{}{}
	expected[9] = struct{}{}
	assert.Equal(t, result, expected)
}

func TestFileIndicesIncorrect(t *testing.T) {
	_, err := util.ParseFileIndices("-1-2")
	if err == nil {
		t.Fatal("no error occured")
	}
}

func TestFileIndicesIncorrect2(t *testing.T) {
	_, err := util.ParseFileIndices(",,")
	if err == nil {
		t.Fatal("no error occured")
	}
}

func TestFileIndicesIncorrect3(t *testing.T) {
	_, err := util.ParseFileIndices("1-2-3")
	if err == nil {
		t.Fatal("no error occured")
	}
}

func TestFileIndicesIncorrect4(t *testing.T) {
	_, err := util.ParseFileIndices("9-8")
	if err == nil {
		t.Fatal("no error occured")
	}
}
