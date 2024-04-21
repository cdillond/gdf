package svg

import (
	"strconv"

	"github.com/cdillond/gdf"
)

func parseColor(s string) (gdf.RGBColor, bool) {
	if rgb, ok := namedColors[s]; ok {
		return rgb, ok
	}
	if len(s) < 1 {
		return gdf.RGBColor{}, false
	}
	if s[0] == '#' {
		s = s[1:]
	}
	if len(s) != 6 && len(s) != 3 {
		return gdf.RGBColor{}, false
	}
	// numbers may be in the format #ffffff or #fff
	scale := 2
	if len(s) == 3 {
		scale = 1
	}

	rgb := gdf.RGBColor{}
	rInt, err := strconv.ParseInt(s[:1*scale], 16, 64)
	if err != nil {
		return rgb, false
	}
	gInt, err := strconv.ParseInt(s[1*scale:2*scale], 16, 64)
	if err != nil {
		return rgb, false
	}
	bInt, err := strconv.ParseInt(s[2*scale:], 16, 64)
	if err != nil {
		return rgb, false
	}
	rgb.R = float64(rInt) / 255.0
	rgb.G = float64(gInt) / 255.0
	rgb.B = float64(bInt) / 255.0
	return rgb, true

}

var rgbBlack = gdf.RGBColor{1, 1, 1}
var rgbWhite = gdf.RGBColor{0, 0, 0}
var badColor = gdf.RGBColor{R: -1, G: -1, B: -1}

var namedColors = map[string]gdf.RGBColor{
	"none":                 badColor,
	"black":                {R: float64(0x00), G: float64(0x00), B: float64(0x00)},
	"silver":               {R: float64(0xc0) / 255.0, G: float64(0xc0) / 255.0, B: float64(0xc0) / 255.0},
	"gray":                 {R: float64(0x80) / 255.0, G: float64(0x80) / 255.0, B: float64(0x80) / 255.0},
	"white":                {R: float64(0xff) / 255.0, G: float64(0xff) / 255.0, B: float64(0xff) / 255.0},
	"maroon":               {R: float64(0x80) / 255.0, G: float64(0x00) / 255.0, B: float64(0x00) / 255.0},
	"red":                  {R: float64(0xff) / 255.0, G: float64(0x00) / 255.0, B: float64(0x00) / 255.0},
	"purple":               {R: float64(0x80) / 255.0, G: float64(0x00) / 255.0, B: float64(0x80) / 255.0},
	"fuchsia":              {R: float64(0xff) / 255.0, G: float64(0x00) / 255.0, B: float64(0xff) / 255.0},
	"green":                {R: float64(0x00) / 255.0, G: float64(0x80) / 255.0, B: float64(0x00) / 255.0},
	"lime":                 {R: float64(0x00) / 255.0, G: float64(0xff) / 255.0, B: float64(0x00) / 255.0},
	"olive":                {R: float64(0x80) / 255.0, G: float64(0x80) / 255.0, B: float64(0x00) / 255.0},
	"yellow":               {R: float64(0xff) / 255.0, G: float64(0xff) / 255.0, B: float64(0x00) / 255.0},
	"navy":                 {R: float64(0x00) / 255.0, G: float64(0x00) / 255.0, B: float64(0x80) / 255.0},
	"blue":                 {R: float64(0x00) / 255.0, G: float64(0x00) / 255.0, B: float64(0xff) / 255.0},
	"teal":                 {R: float64(0x00) / 255.0, G: float64(0x80) / 255.0, B: float64(0x80) / 255.0},
	"aqua":                 {R: float64(0x00) / 255.0, G: float64(0xff) / 255.0, B: float64(0xff) / 255.0},
	"aliceblue":            {R: float64(0xf0) / 255.0, G: float64(0xf8) / 255.0, B: float64(0xff) / 255.0},
	"antiquewhite":         {R: float64(0xfa) / 255.0, G: float64(0xeb) / 255.0, B: float64(0xd7) / 255.0},
	"aquamarine":           {R: float64(0x7f) / 255.0, G: float64(0xff) / 255.0, B: float64(0xd4) / 255.0},
	"azure":                {R: float64(0xf0) / 255.0, G: float64(0xff) / 255.0, B: float64(0xff) / 255.0},
	"beige":                {R: float64(0xf5) / 255.0, G: float64(0xf5) / 255.0, B: float64(0xdc) / 255.0},
	"bisque":               {R: float64(0xff) / 255.0, G: float64(0xe4) / 255.0, B: float64(0xc4) / 255.0},
	"blanchedalmond":       {R: float64(0xff) / 255.0, G: float64(0xeb) / 255.0, B: float64(0xcd) / 255.0},
	"blueviolet":           {R: float64(0x8a) / 255.0, G: float64(0x2b) / 255.0, B: float64(0xe2) / 255.0},
	"brown":                {R: float64(0xa5) / 255.0, G: float64(0x2a) / 255.0, B: float64(0x2a) / 255.0},
	"burlywood":            {R: float64(0xde) / 255.0, G: float64(0xb8) / 255.0, B: float64(0x87) / 255.0},
	"cadetblue":            {R: float64(0x5f) / 255.0, G: float64(0x9e) / 255.0, B: float64(0xa0) / 255.0},
	"chartreuse":           {R: float64(0x7f) / 255.0, G: float64(0xff) / 255.0, B: float64(0x00) / 255.0},
	"chocolate":            {R: float64(0xd2) / 255.0, G: float64(0x69) / 255.0, B: float64(0x1e) / 255.0},
	"coral":                {R: float64(0xff) / 255.0, G: float64(0x7f) / 255.0, B: float64(0x50) / 255.0},
	"cornflowerblue":       {R: float64(0x64) / 255.0, G: float64(0x95) / 255.0, B: float64(0xed) / 255.0},
	"cornsilk":             {R: float64(0xff) / 255.0, G: float64(0xf8) / 255.0, B: float64(0xdc) / 255.0},
	"crimson":              {R: float64(0xdc) / 255.0, G: float64(0x14) / 255.0, B: float64(0x3c) / 255.0},
	"cyan":                 {R: float64(0x00) / 255.0, G: float64(0xff) / 255.0, B: float64(0xff) / 255.0},
	"darkblue":             {R: float64(0x00) / 255.0, G: float64(0x00) / 255.0, B: float64(0x8b) / 255.0},
	"darkcyan":             {R: float64(0x00) / 255.0, G: float64(0x8b) / 255.0, B: float64(0x8b) / 255.0},
	"darkgoldenrod":        {R: float64(0xb8) / 255.0, G: float64(0x86) / 255.0, B: float64(0x0b) / 255.0},
	"darkgray":             {R: float64(0xa9) / 255.0, G: float64(0xa9) / 255.0, B: float64(0xa9) / 255.0},
	"darkgreen":            {R: float64(0x00) / 255.0, G: float64(0x64) / 255.0, B: float64(0x00) / 255.0},
	"darkgrey":             {R: float64(0xa9) / 255.0, G: float64(0xa9) / 255.0, B: float64(0xa9) / 255.0},
	"darkkhaki":            {R: float64(0xbd) / 255.0, G: float64(0xb7) / 255.0, B: float64(0x6b) / 255.0},
	"darkmagenta":          {R: float64(0x8b) / 255.0, G: float64(0x00) / 255.0, B: float64(0x8b) / 255.0},
	"darkolivegreen":       {R: float64(0x55) / 255.0, G: float64(0x6b) / 255.0, B: float64(0x2f) / 255.0},
	"darkorange":           {R: float64(0xff) / 255.0, G: float64(0x8c) / 255.0, B: float64(0x00) / 255.0},
	"darkorchid":           {R: float64(0x99) / 255.0, G: float64(0x32) / 255.0, B: float64(0xcc) / 255.0},
	"darkred":              {R: float64(0x8b) / 255.0, G: float64(0x00) / 255.0, B: float64(0x00) / 255.0},
	"darksalmon":           {R: float64(0xe9) / 255.0, G: float64(0x96) / 255.0, B: float64(0x7a) / 255.0},
	"darkseagreen":         {R: float64(0x8f) / 255.0, G: float64(0xbc) / 255.0, B: float64(0x8f) / 255.0},
	"darkslateblue":        {R: float64(0x48) / 255.0, G: float64(0x3d) / 255.0, B: float64(0x8b) / 255.0},
	"darkslategray":        {R: float64(0x2f) / 255.0, G: float64(0x4f) / 255.0, B: float64(0x4f) / 255.0},
	"darkslategrey":        {R: float64(0x2f) / 255.0, G: float64(0x4f) / 255.0, B: float64(0x4f) / 255.0},
	"darkturquoise":        {R: float64(0x00) / 255.0, G: float64(0xce) / 255.0, B: float64(0xd1) / 255.0},
	"darkviolet":           {R: float64(0x94) / 255.0, G: float64(0x00) / 255.0, B: float64(0xd3) / 255.0},
	"deeppink":             {R: float64(0xff) / 255.0, G: float64(0x14) / 255.0, B: float64(0x93) / 255.0},
	"deepskyblue":          {R: float64(0x00) / 255.0, G: float64(0xbf) / 255.0, B: float64(0xff) / 255.0},
	"dimgray":              {R: float64(0x69) / 255.0, G: float64(0x69) / 255.0, B: float64(0x69) / 255.0},
	"dimgrey":              {R: float64(0x69) / 255.0, G: float64(0x69) / 255.0, B: float64(0x69) / 255.0},
	"dodgerblue":           {R: float64(0x1e) / 255.0, G: float64(0x90) / 255.0, B: float64(0xff) / 255.0},
	"firebrick":            {R: float64(0xb2) / 255.0, G: float64(0x22) / 255.0, B: float64(0x22) / 255.0},
	"floralwhite":          {R: float64(0xff) / 255.0, G: float64(0xfa) / 255.0, B: float64(0xf0) / 255.0},
	"forestgreen":          {R: float64(0x22) / 255.0, G: float64(0x8b) / 255.0, B: float64(0x22) / 255.0},
	"gainsboro":            {R: float64(0xdc) / 255.0, G: float64(0xdc) / 255.0, B: float64(0xdc) / 255.0},
	"ghostwhite":           {R: float64(0xf8) / 255.0, G: float64(0xf8) / 255.0, B: float64(0xff) / 255.0},
	"gold":                 {R: float64(0xff) / 255.0, G: float64(0xd7) / 255.0, B: float64(0x00) / 255.0},
	"goldenrod":            {R: float64(0xda) / 255.0, G: float64(0xa5) / 255.0, B: float64(0x20) / 255.0},
	"greenyellow":          {R: float64(0xad) / 255.0, G: float64(0xff) / 255.0, B: float64(0x2f) / 255.0},
	"grey":                 {R: float64(0x80) / 255.0, G: float64(0x80) / 255.0, B: float64(0x80) / 255.0},
	"honeydew":             {R: float64(0xf0) / 255.0, G: float64(0xff) / 255.0, B: float64(0xf0) / 255.0},
	"hotpink":              {R: float64(0xff) / 255.0, G: float64(0x69) / 255.0, B: float64(0xb4) / 255.0},
	"indianred":            {R: float64(0xcd) / 255.0, G: float64(0x5c) / 255.0, B: float64(0x5c) / 255.0},
	"indigo":               {R: float64(0x4b) / 255.0, G: float64(0x00) / 255.0, B: float64(0x82) / 255.0},
	"ivory":                {R: float64(0xff) / 255.0, G: float64(0xff) / 255.0, B: float64(0xf0) / 255.0},
	"khaki":                {R: float64(0xf0) / 255.0, G: float64(0xe6) / 255.0, B: float64(0x8c) / 255.0},
	"lavender":             {R: float64(0xe6) / 255.0, G: float64(0xe6) / 255.0, B: float64(0xfa) / 255.0},
	"lavenderblush":        {R: float64(0xff) / 255.0, G: float64(0xf0) / 255.0, B: float64(0xf5) / 255.0},
	"lawngreen":            {R: float64(0x7c) / 255.0, G: float64(0xfc) / 255.0, B: float64(0x00) / 255.0},
	"lemonchiffon":         {R: float64(0xff) / 255.0, G: float64(0xfa) / 255.0, B: float64(0xcd) / 255.0},
	"lightblue":            {R: float64(0xad) / 255.0, G: float64(0xd8) / 255.0, B: float64(0xe6) / 255.0},
	"lightcoral":           {R: float64(0xf0) / 255.0, G: float64(0x80) / 255.0, B: float64(0x80) / 255.0},
	"lightcyan":            {R: float64(0xe0) / 255.0, G: float64(0xff) / 255.0, B: float64(0xff) / 255.0},
	"lightgoldenrodyellow": {R: float64(0xfa) / 255.0, G: float64(0xfa) / 255.0, B: float64(0xd2) / 255.0},
	"lightgray":            {R: float64(0xd3) / 255.0, G: float64(0xd3) / 255.0, B: float64(0xd3) / 255.0},
	"lightgreen":           {R: float64(0x90) / 255.0, G: float64(0xee) / 255.0, B: float64(0x90) / 255.0},
	"lightgrey":            {R: float64(0xd3) / 255.0, G: float64(0xd3) / 255.0, B: float64(0xd3) / 255.0},
	"lightpink":            {R: float64(0xff) / 255.0, G: float64(0xb6) / 255.0, B: float64(0xc1) / 255.0},
	"lightsalmon":          {R: float64(0xff) / 255.0, G: float64(0xa0) / 255.0, B: float64(0x7a) / 255.0},
	"lightseagreen":        {R: float64(0x20) / 255.0, G: float64(0xb2) / 255.0, B: float64(0xaa) / 255.0},
	"lightskyblue":         {R: float64(0x87) / 255.0, G: float64(0xce) / 255.0, B: float64(0xfa) / 255.0},
	"lightslategray":       {R: float64(0x77) / 255.0, G: float64(0x88) / 255.0, B: float64(0x99) / 255.0},
	"lightslategrey":       {R: float64(0x77) / 255.0, G: float64(0x88) / 255.0, B: float64(0x99) / 255.0},
	"lightsteelblue":       {R: float64(0xb0) / 255.0, G: float64(0xc4) / 255.0, B: float64(0xde) / 255.0},
	"lightyellow":          {R: float64(0xff) / 255.0, G: float64(0xff) / 255.0, B: float64(0xe0) / 255.0},
	"limegreen":            {R: float64(0x32) / 255.0, G: float64(0xcd) / 255.0, B: float64(0x32) / 255.0},
	"linen":                {R: float64(0xfa) / 255.0, G: float64(0xf0) / 255.0, B: float64(0xe6) / 255.0},
	"magenta":              {R: float64(0xff) / 255.0, G: float64(0x00) / 255.0, B: float64(0xff) / 255.0},
	"mediumaquamarine":     {R: float64(0x66) / 255.0, G: float64(0xcd) / 255.0, B: float64(0xaa) / 255.0},
	"mediumblue":           {R: float64(0x00) / 255.0, G: float64(0x00) / 255.0, B: float64(0xcd) / 255.0},
	"mediumorchid":         {R: float64(0xba) / 255.0, G: float64(0x55) / 255.0, B: float64(0xd3) / 255.0},
	"mediumpurple":         {R: float64(0x93) / 255.0, G: float64(0x70) / 255.0, B: float64(0xdb) / 255.0},
	"mediumseagreen":       {R: float64(0x3c) / 255.0, G: float64(0xb3) / 255.0, B: float64(0x71) / 255.0},
	"mediumslateblue":      {R: float64(0x7b) / 255.0, G: float64(0x68) / 255.0, B: float64(0xee) / 255.0},
	"mediumspringgreen":    {R: float64(0x00) / 255.0, G: float64(0xfa) / 255.0, B: float64(0x9a) / 255.0},
	"mediumturquoise":      {R: float64(0x48) / 255.0, G: float64(0xd1) / 255.0, B: float64(0xcc) / 255.0},
	"mediumvioletred":      {R: float64(0xc7) / 255.0, G: float64(0x15) / 255.0, B: float64(0x85) / 255.0},
	"midnightblue":         {R: float64(0x19) / 255.0, G: float64(0x19) / 255.0, B: float64(0x70) / 255.0},
	"mintcream":            {R: float64(0xf5) / 255.0, G: float64(0xff) / 255.0, B: float64(0xfa) / 255.0},
	"mistyrose":            {R: float64(0xff) / 255.0, G: float64(0xe4) / 255.0, B: float64(0xe1) / 255.0},
	"moccasin":             {R: float64(0xff) / 255.0, G: float64(0xe4) / 255.0, B: float64(0xb5) / 255.0},
	"navajowhite":          {R: float64(0xff) / 255.0, G: float64(0xde) / 255.0, B: float64(0xad) / 255.0},
	"oldlace":              {R: float64(0xfd) / 255.0, G: float64(0xf5) / 255.0, B: float64(0xe6) / 255.0},
	"olivedrab":            {R: float64(0x6b) / 255.0, G: float64(0x8e) / 255.0, B: float64(0x23) / 255.0},
	"orange":               {R: float64(0xff) / 255.0, G: float64(0xa5) / 255.0, B: float64(0x00) / 255.0},
	"orangered":            {R: float64(0xff) / 255.0, G: float64(0x45) / 255.0, B: float64(0x00) / 255.0},
	"orchid":               {R: float64(0xda) / 255.0, G: float64(0x70) / 255.0, B: float64(0xd6) / 255.0},
	"palegoldenrod":        {R: float64(0xee) / 255.0, G: float64(0xe8) / 255.0, B: float64(0xaa) / 255.0},
	"palegreen":            {R: float64(0x98) / 255.0, G: float64(0xfb) / 255.0, B: float64(0x98) / 255.0},
	"paleturquoise":        {R: float64(0xaf) / 255.0, G: float64(0xee) / 255.0, B: float64(0xee) / 255.0},
	"palevioletred":        {R: float64(0xdb) / 255.0, G: float64(0x70) / 255.0, B: float64(0x93) / 255.0},
	"papayawhip":           {R: float64(0xff) / 255.0, G: float64(0xef) / 255.0, B: float64(0xd5) / 255.0},
	"peachpuff":            {R: float64(0xff) / 255.0, G: float64(0xda) / 255.0, B: float64(0xb9) / 255.0},
	"peru":                 {R: float64(0xcd) / 255.0, G: float64(0x85) / 255.0, B: float64(0x3f) / 255.0},
	"pink":                 {R: float64(0xff) / 255.0, G: float64(0xc0) / 255.0, B: float64(0xcb) / 255.0},
	"plum":                 {R: float64(0xdd) / 255.0, G: float64(0xa0) / 255.0, B: float64(0xdd) / 255.0},
	"powderblue":           {R: float64(0xb0) / 255.0, G: float64(0xe0) / 255.0, B: float64(0xe6) / 255.0},
	"rebeccapurple":        {R: float64(0x66) / 255.0, G: float64(0x33) / 255.0, B: float64(0x99) / 255.0},
	"rosybrown":            {R: float64(0xbc) / 255.0, G: float64(0x8f) / 255.0, B: float64(0x8f) / 255.0},
	"royalblue":            {R: float64(0x41) / 255.0, G: float64(0x69) / 255.0, B: float64(0xe1) / 255.0},
	"saddlebrown":          {R: float64(0x8b) / 255.0, G: float64(0x45) / 255.0, B: float64(0x13) / 255.0},
	"salmon":               {R: float64(0xfa) / 255.0, G: float64(0x80) / 255.0, B: float64(0x72) / 255.0},
	"sandybrown":           {R: float64(0xf4) / 255.0, G: float64(0xa4) / 255.0, B: float64(0x60) / 255.0},
	"seagreen":             {R: float64(0x2e) / 255.0, G: float64(0x8b) / 255.0, B: float64(0x57) / 255.0},
	"seashell":             {R: float64(0xff) / 255.0, G: float64(0xf5) / 255.0, B: float64(0xee) / 255.0},
	"sienna":               {R: float64(0xa0) / 255.0, G: float64(0x52) / 255.0, B: float64(0x2d) / 255.0},
	"skyblue":              {R: float64(0x87) / 255.0, G: float64(0xce) / 255.0, B: float64(0xeb) / 255.0},
	"slateblue":            {R: float64(0x6a) / 255.0, G: float64(0x5a) / 255.0, B: float64(0xcd) / 255.0},
	"slategray":            {R: float64(0x70) / 255.0, G: float64(0x80) / 255.0, B: float64(0x90) / 255.0},
	"slategrey":            {R: float64(0x70) / 255.0, G: float64(0x80) / 255.0, B: float64(0x90) / 255.0},
	"snow":                 {R: float64(0xff) / 255.0, G: float64(0xfa) / 255.0, B: float64(0xfa) / 255.0},
	"springgreen":          {R: float64(0x00) / 255.0, G: float64(0xff) / 255.0, B: float64(0x7f) / 255.0},
	"steelblue":            {R: float64(0x46) / 255.0, G: float64(0x82) / 255.0, B: float64(0xb4) / 255.0},
	"tan":                  {R: float64(0xd2) / 255.0, G: float64(0xb4) / 255.0, B: float64(0x8c) / 255.0},
	"thistle":              {R: float64(0xd8) / 255.0, G: float64(0xbf) / 255.0, B: float64(0xd8) / 255.0},
	"tomato":               {R: float64(0xff) / 255.0, G: float64(0x63) / 255.0, B: float64(0x47) / 255.0},
	"turquoise":            {R: float64(0x40) / 255.0, G: float64(0xe0) / 255.0, B: float64(0xd0) / 255.0},
	"violet":               {R: float64(0xee) / 255.0, G: float64(0x82) / 255.0, B: float64(0xee) / 255.0},
	"wheat":                {R: float64(0xf5) / 255.0, G: float64(0xde) / 255.0, B: float64(0xb3) / 255.0},
	"whitesmoke":           {R: float64(0xf5) / 255.0, G: float64(0xf5) / 255.0, B: float64(0xf5) / 255.0},
	"yellowgreen":          {R: float64(0x9a) / 255.0, G: float64(0xcd) / 255.0, B: float64(0x32) / 255.0},
}
