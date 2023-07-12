package util

import (
	"log"
	"reflect"
)

// AssertEqual checks if a equal b
func AssertEqual(a, b any) {
	if !reflect.DeepEqual(a, b) {
		log.Fatalf("ASSERTION: %+v != %+v", a, b)
	}
}
