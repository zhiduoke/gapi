package jtop

import (
	"testing"
)

const data = "{\"animals\":{\"dog\":[{\"name\":\"Rufus\",\"age\":15},{\"name\":\"Marty\",\"age\":null}]}}"

func TestIter_Consume(t *testing.T) {
	iter := NewIter([]byte(data))
	for iter.Next() {
		token := iter.Consume()
		t.Logf("Kind: %d ,Value: %s", token.Kind, token.Value)
		if token.Kind == Invalid {
			break
		}
	}
}
