package content

import (
	"fmt"

	"github.com/cdillond/gdf"
)

type Table struct {
	NR, NC     int // number of rows, number of columns
	Rows       [][]gdf.ShapedRun
	RowHeights []float64
	ColWidths  []float64
}

func NewTable(height, width float64, numRows, numCols int) Table {
	rhs := make([]float64, numRows)
	cws := make([]float64, numCols)
	for i := range rhs {
		rhs[i] = height / float64(numRows)
	}
	for i := range cws {
		cws[i] = width / float64(numCols)
	}
	return Table{
		NR:         numRows,
		NC:         numCols,
		RowHeights: rhs,
		ColWidths:  cws,
	}
}

func DrawTable(cs *gdf.ContentStream, t Table) error {
	if cs.TextObj == nil {
		return fmt.Errorf("you must initialize a text object before drawing a table")
	}
	// first draw text

	start := cs.TextCursor()
	cs.TStar()
	var hsum float64
	for i, tr := range t.Rows {
		if i != 0 {
			cs.Tm(gdf.Translate(gdf.NewMatrix(), start.X, start.Y-hsum-cs.Leading))
		}

		for j, td := range tr {
			for k := len(td.Txt); k >= 0; k-- {
				if t.ColWidths[i] > cs.ShapedTextExtentPts(string(td.Txt[:k])) {
					dif := gdf.CenterH(gdf.Shape(string(td.Txt[:k]), cs.Font), gdf.Rect{0, 0, t.ColWidths[i], 0}, cs.FontSize)
					cs.Td(dif, 0)
					cs.TJSpace(td.Txt[:k], td.Kerns[:k], 0)
					cs.Td(-dif, 0)
					break
				}
			}
			cs.TStar()
			cs.Td(t.ColWidths[j], cs.Leading)
		}
		hsum += t.RowHeights[i]
	}

	// then draw table borders
	var width, height float64
	for i := 0; i < t.NR; i++ {
		height += t.RowHeights[i]
	}
	for i := 0; i < t.NC; i++ {
		width += t.ColWidths[i]
	}
	cs.Re(start.X, start.Y-height, width, height)
	cs.Stroke()
	var curX, curY float64
	cs.WSmall(2)
	for i := 0; i < t.NR-1; i++ {
		curY += t.RowHeights[i]
		cs.MoveTo(start.X, start.Y-curY)
		cs.LineTo(start.X+width, start.Y-curY)
		if i == 1 {
			cs.WSmall(1)
		}
		cs.Stroke()

	}
	for i := 0; i < t.NC-1; i++ {
		curX += t.ColWidths[i]
		cs.MoveTo(start.X+curX, start.Y-height)
		cs.LineTo(start.X+curX, start.Y)
		cs.Stroke()
	}

	return nil
}
