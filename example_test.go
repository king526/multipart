// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multipart_test

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/king526/multipart"
)

func Test_1(t *testing.T) {
	b := multipart.NewFormBody()
	fmt.Println(b.WriteField("f1", "v1"))
	fmt.Println(b.WriteField("f2", "v2"))
	fmt.Println(b.CreateFromByPath("f4", "f44", "example_test.go"))
	fmt.Println(b.WriteField("f3", "v3"))

	bs, err := ioutil.ReadAll(b)
	t.Log(err, string(bs))
}
