package svg

import "github.com/cdillond/gdf"

func WriteCmd(c *gdf.ContentStream, cmd pdfPathCmd) {
	switch cmd.op {
	case moveTo:
		c.MoveTo(cmd.args[0].X, cmd.args[0].Y)
	case lineTo:
		c.LineTo(cmd.args[0].X, cmd.args[0].Y)
	case curveTo:
		c.CubicBezier1(cmd.args[0].X, cmd.args[0].Y, cmd.args[1].X, cmd.args[1].Y, cmd.args[2].X, cmd.args[2].Y)
	case closePath:
		c.ClosePath()
	case circle:
		c.Circle(cmd.args[0].X, cmd.args[0].Y, cmd.args[1].X)
	case ellipse:
		c.Ellipse(cmd.args[0].X, cmd.args[0].Y, cmd.args[1].X, cmd.args[1].Y)
	}
}
