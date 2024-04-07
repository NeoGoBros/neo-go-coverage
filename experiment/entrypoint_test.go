package experiment;

import (
	"testing"
)

var (
	validKey   = []byte{1, 2, 3, 4, 5}
	invalidKey = []byte{1, 2, 3}
)

func Test1(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Errorf("The code did not panic")
        }
    }()

    GetNumber(validKey)
}

func Test2(t *testing.T) {
	PutNumber(validKey, 42)
	GetNumber(validKey)
}

func Test3(t *testing.T) {
	defer func() {
        if r := recover(); r == nil {
            t.Errorf("The code did not panic")
        }
    }()

	PutNumber(invalidKey, 42);
}

func Test4(t *testing.T) {
	defer func() {
        if r := recover(); r == nil {
            t.Errorf("The code did not panic")
        }
    }()

	GetNumber(invalidKey);
}
