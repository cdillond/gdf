package gdf

import "io"

/*
The obj interface represents a node in the PDF document graph. With the exception of the root node,
objs must have at least 1 parent. Objs not reachable from the root node are not included in the final PDF document.
When PDF.WriteTo(w) is called, the PDF graph is built recursively by the following procedure:

 1. call obj.mark() on a node and append it to the slice of document objects

 2. call obj.children()

    a. for each unvisited child node, goto: step 1

    b. if obj.children() returns nil, return

After all reachable nodes have been added to the document graph, they are sequentially written to the destination
io.Writer. Calls to obj.encode(w io.Writer) should write the PDF-encoded byte-representation of obj to w and return
the number of bytes written. Calls to id() should return 0 if obj.mark(i) has not yet been invoked, and i if it has.
The returned int represents the obj's handle, by which it can be indirectly referenced from elsewhere in the PDF document graph.
obj.id() should ONLY be invoked during the execution of an obj.encode() method or during the process of building the document graph.
*/
type obj interface {
	mark(i int)
	children() []obj
	id() int
	encode(w io.Writer) (int, error)
}
