package gdf

import "io"

type obj interface {
	setRef(i int)
	refNum() int
	children() []obj
	encode(w io.Writer) (int, error)
}

type Name string

func (n Name) String() string { return "/" + string(n) }
