package resolvers

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructFieldCapsNameMapReturnErrorForNonStruct(t *testing.T) {
	tt := 1
	_, err := StructFieldCapsNameMap(&tt)
	assert.Equal(t, err, errors.New("must provide a struct"))
}

func TestStructFieldCapsNameMap(t *testing.T) {
	type TestType struct {
		id   int
		name string
	}
	tt := TestType{
		id:   1,
		name: "name",
	}
	m, err := StructFieldCapsNameMap(&tt)
	assert.Nil(t, err)
	expected := map[string]int{
		"ID":   0,
		"NAME": 1,
	}
	fmt.Printf("%+v", m)
	assert.True(t, reflect.DeepEqual(m, expected))
}
