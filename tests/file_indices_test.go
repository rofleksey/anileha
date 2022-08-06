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
	expectedValues := make(map[int]struct{}, 10)
	expectedValues[0] = struct{}{}
	expectedValues[1] = struct{}{}
	expectedValues[2] = struct{}{}
	expectedValues[3] = struct{}{}
	expectedValues[4] = struct{}{}
	expectedValues[6] = struct{}{}
	expectedValues[8] = struct{}{}
	expectedValues[9] = struct{}{}
	expected := util.FileIndices{
		Values: expectedValues,
	}
	assert.Equal(t, result, expected)
}

func TestFileIndicesInfinite(t *testing.T) {
	result, err := util.ParseFileIndices("*")
	if err != nil {
		t.Fatal(err)
	}
	expected := util.FileIndices{
		Infinite: true,
	}
	assert.Equal(t, result, expected)
	assert.Equal(t, result.Contains(0), true)
	assert.Equal(t, result.Contains(10), true)
	assert.Equal(t, result.Contains(100), true)
	assert.Equal(t, result.Contains(1000), true)
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
