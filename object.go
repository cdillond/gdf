package gdf

import "io"

type Obj interface {
	SetRef(i int)
	RefNum() int
	Children() []Obj
	Encode(w io.Writer) (int, error)
}

type Name string

func ToString(n Name) string { return "/" + string(n) }
