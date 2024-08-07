package svg

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/cdillond/gdf"
)

func parseColor(s string) (c cl, o opacity) {
	if s == "none" {
		c.isSet = true
		c.isNone = true
		return c, o
	}
	if s == "transparent" {
		o.isSet = true
		o.val = 0
		return c, o
	}
	if len(s) > 4 && s[:4] == "rgb(" {
		return parseRGBFunc(s)
	}

	if rgb, ok := namedColors[s]; ok {
		c.isSet = true
		c.RGBColor = rgb
		return c, o
	}

	if len(s) < 1 {
		return c, o
	}
	if s[0] == '#' {
		s = s[1:]
	}
	if len(s) != 6 && len(s) != 3 {
		return c, o
	}
	// numbers may be in the format #ffffff or #fff
	scale := 2
	if len(s) == 3 {
		scale = 1
	}

	rInt, err := strconv.ParseInt(s[:1*scale], 16, 64)
	if err != nil {
		return c, o
	}
	gInt, err := strconv.ParseInt(s[1*scale:2*scale], 16, 64)
	if err != nil {
		return c, o
	}
	bInt, err := strconv.ParseInt(s[2*scale:], 16, 64)
	if err != nil {
		return c, o
	}
	c.isSet = true
	if scale == 2 {
		c.RGBColor.R = float64(rInt) / 255.0
		c.RGBColor.G = float64(gInt) / 255.0
		c.RGBColor.B = float64(bInt) / 255.0
	} else {
		c.RGBColor.R = float64(rInt) / 16.0
		c.RGBColor.G = float64(gInt) / 16.0
		c.RGBColor.B = float64(bInt) / 16.0
	}
	return c, o

}

func parseRGBFunc(s string) (c cl, o opacity) {
	s = s[4:]
	s = strings.Trim(s, "()\u0020\n\r\t\v\f\u0085\u00A0")

	n := strings.IndexByte(s, '/')
	if n > 0 && n < len(s)-1 {
		after := s[n+1:]
		s = s[:n]
		after = strings.TrimSpace(after)
		num, _ := parseNumPct(after)
		o.isSet = true
		o.val = num
	}
	cols := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == ','
	})
	// this is an ad-hoc parsing method which does not account for edge cases or alpha values, e.g. rgb(127 255 127 / 80% )
	if len(cols) < 3 {
		return c, o
	}
	var err error
	var cArr [3]float64
	for i := 0; i < 3; i++ {
		if cols[i] == "none" {
			continue
		} else if strings.Contains(cols[i], "%") {
			cArr[i], err = strconv.ParseFloat(cols[i][:len(cols[i])-1], 64)
			cArr[i] /= 100.
		} else {
			cArr[i], err = strconv.ParseFloat(cols[i], 64)
			cArr[i] /= 255.
		}
		if err != nil {
			return c, o
		}
	}
	c.isSet = true
	c.RGBColor = gdf.RGBColor{cArr[0], cArr[1], cArr[2]}
	return c, o
}

var namedColors = map[string]gdf.RGBColor{
	"black":                {float64(0x00), float64(0x00), float64(0x00)},
	"silver":               {float64(0xc0) / 255.0, float64(0xc0) / 255.0, float64(0xc0) / 255.0},
	"gray":                 {float64(0x80) / 255.0, float64(0x80) / 255.0, float64(0x80) / 255.0},
	"white":                {float64(0xff) / 255.0, float64(0xff) / 255.0, float64(0xff) / 255.0},
	"maroon":               {float64(0x80) / 255.0, float64(0x00) / 255.0, float64(0x00) / 255.0},
	"red":                  {float64(0xff) / 255.0, float64(0x00) / 255.0, float64(0x00) / 255.0},
	"purple":               {float64(0x80) / 255.0, float64(0x00) / 255.0, float64(0x80) / 255.0},
	"fuchsia":              {float64(0xff) / 255.0, float64(0x00) / 255.0, float64(0xff) / 255.0},
	"green":                {float64(0x00) / 255.0, float64(0x80) / 255.0, float64(0x00) / 255.0},
	"lime":                 {float64(0x00) / 255.0, float64(0xff) / 255.0, float64(0x00) / 255.0},
	"olive":                {float64(0x80) / 255.0, float64(0x80) / 255.0, float64(0x00) / 255.0},
	"yellow":               {float64(0xff) / 255.0, float64(0xff) / 255.0, float64(0x00) / 255.0},
	"navy":                 {float64(0x00) / 255.0, float64(0x00) / 255.0, float64(0x80) / 255.0},
	"blue":                 {float64(0x00) / 255.0, float64(0x00) / 255.0, float64(0xff) / 255.0},
	"teal":                 {float64(0x00) / 255.0, float64(0x80) / 255.0, float64(0x80) / 255.0},
	"aqua":                 {float64(0x00) / 255.0, float64(0xff) / 255.0, float64(0xff) / 255.0},
	"aliceblue":            {float64(0xf0) / 255.0, float64(0xf8) / 255.0, float64(0xff) / 255.0},
	"antiquewhite":         {float64(0xfa) / 255.0, float64(0xeb) / 255.0, float64(0xd7) / 255.0},
	"aquamarine":           {float64(0x7f) / 255.0, float64(0xff) / 255.0, float64(0xd4) / 255.0},
	"azure":                {float64(0xf0) / 255.0, float64(0xff) / 255.0, float64(0xff) / 255.0},
	"beige":                {float64(0xf5) / 255.0, float64(0xf5) / 255.0, float64(0xdc) / 255.0},
	"bisque":               {float64(0xff) / 255.0, float64(0xe4) / 255.0, float64(0xc4) / 255.0},
	"blanchedalmond":       {float64(0xff) / 255.0, float64(0xeb) / 255.0, float64(0xcd) / 255.0},
	"blueviolet":           {float64(0x8a) / 255.0, float64(0x2b) / 255.0, float64(0xe2) / 255.0},
	"brown":                {float64(0xa5) / 255.0, float64(0x2a) / 255.0, float64(0x2a) / 255.0},
	"burlywood":            {float64(0xde) / 255.0, float64(0xb8) / 255.0, float64(0x87) / 255.0},
	"cadetblue":            {float64(0x5f) / 255.0, float64(0x9e) / 255.0, float64(0xa0) / 255.0},
	"chartreuse":           {float64(0x7f) / 255.0, float64(0xff) / 255.0, float64(0x00) / 255.0},
	"chocolate":            {float64(0xd2) / 255.0, float64(0x69) / 255.0, float64(0x1e) / 255.0},
	"coral":                {float64(0xff) / 255.0, float64(0x7f) / 255.0, float64(0x50) / 255.0},
	"cornflowerblue":       {float64(0x64) / 255.0, float64(0x95) / 255.0, float64(0xed) / 255.0},
	"cornsilk":             {float64(0xff) / 255.0, float64(0xf8) / 255.0, float64(0xdc) / 255.0},
	"crimson":              {float64(0xdc) / 255.0, float64(0x14) / 255.0, float64(0x3c) / 255.0},
	"cyan":                 {float64(0x00) / 255.0, float64(0xff) / 255.0, float64(0xff) / 255.0},
	"darkblue":             {float64(0x00) / 255.0, float64(0x00) / 255.0, float64(0x8b) / 255.0},
	"darkcyan":             {float64(0x00) / 255.0, float64(0x8b) / 255.0, float64(0x8b) / 255.0},
	"darkgoldenrod":        {float64(0xb8) / 255.0, float64(0x86) / 255.0, float64(0x0b) / 255.0},
	"darkgray":             {float64(0xa9) / 255.0, float64(0xa9) / 255.0, float64(0xa9) / 255.0},
	"darkgreen":            {float64(0x00) / 255.0, float64(0x64) / 255.0, float64(0x00) / 255.0},
	"darkgrey":             {float64(0xa9) / 255.0, float64(0xa9) / 255.0, float64(0xa9) / 255.0},
	"darkkhaki":            {float64(0xbd) / 255.0, float64(0xb7) / 255.0, float64(0x6b) / 255.0},
	"darkmagenta":          {float64(0x8b) / 255.0, float64(0x00) / 255.0, float64(0x8b) / 255.0},
	"darkolivegreen":       {float64(0x55) / 255.0, float64(0x6b) / 255.0, float64(0x2f) / 255.0},
	"darkorange":           {float64(0xff) / 255.0, float64(0x8c) / 255.0, float64(0x00) / 255.0},
	"darkorchid":           {float64(0x99) / 255.0, float64(0x32) / 255.0, float64(0xcc) / 255.0},
	"darkred":              {float64(0x8b) / 255.0, float64(0x00) / 255.0, float64(0x00) / 255.0},
	"darksalmon":           {float64(0xe9) / 255.0, float64(0x96) / 255.0, float64(0x7a) / 255.0},
	"darkseagreen":         {float64(0x8f) / 255.0, float64(0xbc) / 255.0, float64(0x8f) / 255.0},
	"darkslateblue":        {float64(0x48) / 255.0, float64(0x3d) / 255.0, float64(0x8b) / 255.0},
	"darkslategray":        {float64(0x2f) / 255.0, float64(0x4f) / 255.0, float64(0x4f) / 255.0},
	"darkslategrey":        {float64(0x2f) / 255.0, float64(0x4f) / 255.0, float64(0x4f) / 255.0},
	"darkturquoise":        {float64(0x00) / 255.0, float64(0xce) / 255.0, float64(0xd1) / 255.0},
	"darkviolet":           {float64(0x94) / 255.0, float64(0x00) / 255.0, float64(0xd3) / 255.0},
	"deeppink":             {float64(0xff) / 255.0, float64(0x14) / 255.0, float64(0x93) / 255.0},
	"deepskyblue":          {float64(0x00) / 255.0, float64(0xbf) / 255.0, float64(0xff) / 255.0},
	"dimgray":              {float64(0x69) / 255.0, float64(0x69) / 255.0, float64(0x69) / 255.0},
	"dimgrey":              {float64(0x69) / 255.0, float64(0x69) / 255.0, float64(0x69) / 255.0},
	"dodgerblue":           {float64(0x1e) / 255.0, float64(0x90) / 255.0, float64(0xff) / 255.0},
	"firebrick":            {float64(0xb2) / 255.0, float64(0x22) / 255.0, float64(0x22) / 255.0},
	"floralwhite":          {float64(0xff) / 255.0, float64(0xfa) / 255.0, float64(0xf0) / 255.0},
	"forestgreen":          {float64(0x22) / 255.0, float64(0x8b) / 255.0, float64(0x22) / 255.0},
	"gainsboro":            {float64(0xdc) / 255.0, float64(0xdc) / 255.0, float64(0xdc) / 255.0},
	"ghostwhite":           {float64(0xf8) / 255.0, float64(0xf8) / 255.0, float64(0xff) / 255.0},
	"gold":                 {float64(0xff) / 255.0, float64(0xd7) / 255.0, float64(0x00) / 255.0},
	"goldenrod":            {float64(0xda) / 255.0, float64(0xa5) / 255.0, float64(0x20) / 255.0},
	"greenyellow":          {float64(0xad) / 255.0, float64(0xff) / 255.0, float64(0x2f) / 255.0},
	"grey":                 {float64(0x80) / 255.0, float64(0x80) / 255.0, float64(0x80) / 255.0},
	"honeydew":             {float64(0xf0) / 255.0, float64(0xff) / 255.0, float64(0xf0) / 255.0},
	"hotpink":              {float64(0xff) / 255.0, float64(0x69) / 255.0, float64(0xb4) / 255.0},
	"indianred":            {float64(0xcd) / 255.0, float64(0x5c) / 255.0, float64(0x5c) / 255.0},
	"indigo":               {float64(0x4b) / 255.0, float64(0x00) / 255.0, float64(0x82) / 255.0},
	"ivory":                {float64(0xff) / 255.0, float64(0xff) / 255.0, float64(0xf0) / 255.0},
	"khaki":                {float64(0xf0) / 255.0, float64(0xe6) / 255.0, float64(0x8c) / 255.0},
	"lavender":             {float64(0xe6) / 255.0, float64(0xe6) / 255.0, float64(0xfa) / 255.0},
	"lavenderblush":        {float64(0xff) / 255.0, float64(0xf0) / 255.0, float64(0xf5) / 255.0},
	"lawngreen":            {float64(0x7c) / 255.0, float64(0xfc) / 255.0, float64(0x00) / 255.0},
	"lemonchiffon":         {float64(0xff) / 255.0, float64(0xfa) / 255.0, float64(0xcd) / 255.0},
	"lightblue":            {float64(0xad) / 255.0, float64(0xd8) / 255.0, float64(0xe6) / 255.0},
	"lightcoral":           {float64(0xf0) / 255.0, float64(0x80) / 255.0, float64(0x80) / 255.0},
	"lightcyan":            {float64(0xe0) / 255.0, float64(0xff) / 255.0, float64(0xff) / 255.0},
	"lightgoldenrodyellow": {float64(0xfa) / 255.0, float64(0xfa) / 255.0, float64(0xd2) / 255.0},
	"lightgray":            {float64(0xd3) / 255.0, float64(0xd3) / 255.0, float64(0xd3) / 255.0},
	"lightgreen":           {float64(0x90) / 255.0, float64(0xee) / 255.0, float64(0x90) / 255.0},
	"lightgrey":            {float64(0xd3) / 255.0, float64(0xd3) / 255.0, float64(0xd3) / 255.0},
	"lightpink":            {float64(0xff) / 255.0, float64(0xb6) / 255.0, float64(0xc1) / 255.0},
	"lightsalmon":          {float64(0xff) / 255.0, float64(0xa0) / 255.0, float64(0x7a) / 255.0},
	"lightseagreen":        {float64(0x20) / 255.0, float64(0xb2) / 255.0, float64(0xaa) / 255.0},
	"lightskyblue":         {float64(0x87) / 255.0, float64(0xce) / 255.0, float64(0xfa) / 255.0},
	"lightslategray":       {float64(0x77) / 255.0, float64(0x88) / 255.0, float64(0x99) / 255.0},
	"lightslategrey":       {float64(0x77) / 255.0, float64(0x88) / 255.0, float64(0x99) / 255.0},
	"lightsteelblue":       {float64(0xb0) / 255.0, float64(0xc4) / 255.0, float64(0xde) / 255.0},
	"lightyellow":          {float64(0xff) / 255.0, float64(0xff) / 255.0, float64(0xe0) / 255.0},
	"limegreen":            {float64(0x32) / 255.0, float64(0xcd) / 255.0, float64(0x32) / 255.0},
	"linen":                {float64(0xfa) / 255.0, float64(0xf0) / 255.0, float64(0xe6) / 255.0},
	"magenta":              {float64(0xff) / 255.0, float64(0x00) / 255.0, float64(0xff) / 255.0},
	"mediumaquamarine":     {float64(0x66) / 255.0, float64(0xcd) / 255.0, float64(0xaa) / 255.0},
	"mediumblue":           {float64(0x00) / 255.0, float64(0x00) / 255.0, float64(0xcd) / 255.0},
	"mediumorchid":         {float64(0xba) / 255.0, float64(0x55) / 255.0, float64(0xd3) / 255.0},
	"mediumpurple":         {float64(0x93) / 255.0, float64(0x70) / 255.0, float64(0xdb) / 255.0},
	"mediumseagreen":       {float64(0x3c) / 255.0, float64(0xb3) / 255.0, float64(0x71) / 255.0},
	"mediumslateblue":      {float64(0x7b) / 255.0, float64(0x68) / 255.0, float64(0xee) / 255.0},
	"mediumspringgreen":    {float64(0x00) / 255.0, float64(0xfa) / 255.0, float64(0x9a) / 255.0},
	"mediumturquoise":      {float64(0x48) / 255.0, float64(0xd1) / 255.0, float64(0xcc) / 255.0},
	"mediumvioletred":      {float64(0xc7) / 255.0, float64(0x15) / 255.0, float64(0x85) / 255.0},
	"midnightblue":         {float64(0x19) / 255.0, float64(0x19) / 255.0, float64(0x70) / 255.0},
	"mintcream":            {float64(0xf5) / 255.0, float64(0xff) / 255.0, float64(0xfa) / 255.0},
	"mistyrose":            {float64(0xff) / 255.0, float64(0xe4) / 255.0, float64(0xe1) / 255.0},
	"moccasin":             {float64(0xff) / 255.0, float64(0xe4) / 255.0, float64(0xb5) / 255.0},
	"navajowhite":          {float64(0xff) / 255.0, float64(0xde) / 255.0, float64(0xad) / 255.0},
	"oldlace":              {float64(0xfd) / 255.0, float64(0xf5) / 255.0, float64(0xe6) / 255.0},
	"olivedrab":            {float64(0x6b) / 255.0, float64(0x8e) / 255.0, float64(0x23) / 255.0},
	"orange":               {float64(0xff) / 255.0, float64(0xa5) / 255.0, float64(0x00) / 255.0},
	"orangered":            {float64(0xff) / 255.0, float64(0x45) / 255.0, float64(0x00) / 255.0},
	"orchid":               {float64(0xda) / 255.0, float64(0x70) / 255.0, float64(0xd6) / 255.0},
	"palegoldenrod":        {float64(0xee) / 255.0, float64(0xe8) / 255.0, float64(0xaa) / 255.0},
	"palegreen":            {float64(0x98) / 255.0, float64(0xfb) / 255.0, float64(0x98) / 255.0},
	"paleturquoise":        {float64(0xaf) / 255.0, float64(0xee) / 255.0, float64(0xee) / 255.0},
	"palevioletred":        {float64(0xdb) / 255.0, float64(0x70) / 255.0, float64(0x93) / 255.0},
	"papayawhip":           {float64(0xff) / 255.0, float64(0xef) / 255.0, float64(0xd5) / 255.0},
	"peachpuff":            {float64(0xff) / 255.0, float64(0xda) / 255.0, float64(0xb9) / 255.0},
	"peru":                 {float64(0xcd) / 255.0, float64(0x85) / 255.0, float64(0x3f) / 255.0},
	"pink":                 {float64(0xff) / 255.0, float64(0xc0) / 255.0, float64(0xcb) / 255.0},
	"plum":                 {float64(0xdd) / 255.0, float64(0xa0) / 255.0, float64(0xdd) / 255.0},
	"powderblue":           {float64(0xb0) / 255.0, float64(0xe0) / 255.0, float64(0xe6) / 255.0},
	"rebeccapurple":        {float64(0x66) / 255.0, float64(0x33) / 255.0, float64(0x99) / 255.0},
	"rosybrown":            {float64(0xbc) / 255.0, float64(0x8f) / 255.0, float64(0x8f) / 255.0},
	"royalblue":            {float64(0x41) / 255.0, float64(0x69) / 255.0, float64(0xe1) / 255.0},
	"saddlebrown":          {float64(0x8b) / 255.0, float64(0x45) / 255.0, float64(0x13) / 255.0},
	"salmon":               {float64(0xfa) / 255.0, float64(0x80) / 255.0, float64(0x72) / 255.0},
	"sandybrown":           {float64(0xf4) / 255.0, float64(0xa4) / 255.0, float64(0x60) / 255.0},
	"seagreen":             {float64(0x2e) / 255.0, float64(0x8b) / 255.0, float64(0x57) / 255.0},
	"seashell":             {float64(0xff) / 255.0, float64(0xf5) / 255.0, float64(0xee) / 255.0},
	"sienna":               {float64(0xa0) / 255.0, float64(0x52) / 255.0, float64(0x2d) / 255.0},
	"skyblue":              {float64(0x87) / 255.0, float64(0xce) / 255.0, float64(0xeb) / 255.0},
	"slateblue":            {float64(0x6a) / 255.0, float64(0x5a) / 255.0, float64(0xcd) / 255.0},
	"slategray":            {float64(0x70) / 255.0, float64(0x80) / 255.0, float64(0x90) / 255.0},
	"slategrey":            {float64(0x70) / 255.0, float64(0x80) / 255.0, float64(0x90) / 255.0},
	"snow":                 {float64(0xff) / 255.0, float64(0xfa) / 255.0, float64(0xfa) / 255.0},
	"springgreen":          {float64(0x00) / 255.0, float64(0xff) / 255.0, float64(0x7f) / 255.0},
	"steelblue":            {float64(0x46) / 255.0, float64(0x82) / 255.0, float64(0xb4) / 255.0},
	"tan":                  {float64(0xd2) / 255.0, float64(0xb4) / 255.0, float64(0x8c) / 255.0},
	"thistle":              {float64(0xd8) / 255.0, float64(0xbf) / 255.0, float64(0xd8) / 255.0},
	"tomato":               {float64(0xff) / 255.0, float64(0x63) / 255.0, float64(0x47) / 255.0},
	"turquoise":            {float64(0x40) / 255.0, float64(0xe0) / 255.0, float64(0xd0) / 255.0},
	"violet":               {float64(0xee) / 255.0, float64(0x82) / 255.0, float64(0xee) / 255.0},
	"wheat":                {float64(0xf5) / 255.0, float64(0xde) / 255.0, float64(0xb3) / 255.0},
	"whitesmoke":           {float64(0xf5) / 255.0, float64(0xf5) / 255.0, float64(0xf5) / 255.0},
	"yellowgreen":          {float64(0x9a) / 255.0, float64(0xcd) / 255.0, float64(0x32) / 255.0},
}
