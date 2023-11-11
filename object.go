package gdf

import "io"

type Obj interface {
	setRef(i int)
	refNum() int
	children() []Obj
	encode(w io.Writer) (int, error)
}

type Name string

func ToString(n Name) string { return "/" + string(n) }
