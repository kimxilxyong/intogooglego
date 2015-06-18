// This is a demo to show how to convert from a normal struct to a reflection type
// and back to a struct without knowing the original one.
// Post is passed as a reflect.Type and the output will be a struct which is
// identical to it (including the embedded Comment struct)

package main

import (
	"errors"
	"fmt"
	"os"
	"reflect"
)

// Ignore the tags in this example, im just too lazy to remove them here
// holds a single post
// You can use ether db or gorp as tag
type Post struct {
	Id       uint64
	Title    string
	Comments []*Comment
}

// holds a single comment bound to a post
type Comment struct {
	Id     uint64
	PostId uint64
	Body   string
}

func CreateAndFillSlice(i interface{}, sliceName string) (interface{}, error) {

	t := reflect.TypeOf(i)
	// Check if the input is a pointer and dereference it if yes
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if the input is a struct
	if t.Kind() != reflect.Struct {
		return nil, errors.New("Input param is not a struct")
	}
	fmt.Printf("Input Type %v:\n", t)

	// Create a new Value from the input type
	// this will be returned to the caller
	v := reflect.New(t).Elem()

	// Get the field named "sliceName" from the input struct, which should be a slice
	s := v.FieldByName(sliceName)
	if s.Kind() == reflect.Slice {

		st := s.Type()
		fmt.Printf("Slice Type %s:\n", st)

		// Get the type of a single slice element
		sliceType := st.Elem()
		// Pointer?
		if sliceType.Kind() == reflect.Ptr {
			// Then dereference it
			sliceType = sliceType.Elem()
		}
		fmt.Printf("Slice Elem Type %v:\n", sliceType)

		for i := 0; i < 5; i++ {
			// Create a new slice element
			newitem := reflect.New(sliceType)
			// Set some field in it
			newitem.Elem().FieldByName("Body").SetString(fmt.Sprintf("XYZ %d", i))
			newitem.Elem().FieldByName("PostId").SetUint(uint64(i * 2))

			// This is the important part here - append and set
			// Append the newitem to the slice in "v" which will be the output
			s.Set(reflect.Append(s, newitem))
		}
	} else {
		return nil, fmt.Errorf("Field %s is not a slice\n", sliceName)
	}

	// IMPORTANT
	// Cast back to the empty interface type
	// So the cast back to Post outside will work
	return v.Interface(), nil
}

func main() {
	var err error
	p := Post{Id: 1, Title: "Title 1"}

	result, err := CreateAndFillSlice(p, "Comments")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	// Cast the returned interface to a Post
	post := result.(Post)
	for i, c := range post.Comments {
		fmt.Printf("Comment %d, Body %s, PostId %d\n", i, c.Body, c.PostId)
	}
}
