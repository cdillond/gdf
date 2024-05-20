package gdf

// graphics state ops (Table 56)
const (
	op_q  = "q\n"  // push graphics state
	op_Q  = "Q\n"  // pop graphics state
	op_cm = "cm\n" // concat matrix
	op_w  = "w\n"  // line width
	op_J  = "J\n"  // line cap
	op_j  = "j\n"  // line join
	op_M  = "M\n"  // miter limit
	op_d  = "d\n"  // dash pattern
	op_ri = "ri\n" // rendering intent
	op_i  = "i\n"  // flatness
	op_gs = "gs\n" // extended graphics state
)

// path construction ops (Table 58)
const (
	op_m  = "m\n"  // begin subpath
	op_l  = "l\n"  // append line
	op_c  = "c\n"  // append cubic Bézier curve 1
	op_v  = "v\n"  // append cubic Bézier curve 2
	op_y  = "y\n"  // append cubic Bézier curve 3
	op_h  = "h\n"  // close subpath
	op_re = "re\n" // append rectangle
)

// path painting ops (Table 59)
const (
	op_S   = "S\n"  // stroke path
	op_s   = "s\n"  // close and stroke path
	op_f   = "f\n"  // fill path non-zero winding
	op_f_X = "f*\n" // fill path even-odd
	op_B   = "B\n"  // fill path non-zero winding and then stroke path
	op_B_X = "B*\n" // fill path even-odd and then stroke path
	op_b   = "b\n"  // close path, fill path non-zero winding, and then stroke path
	op_b_X = "b*\n" // close path, fill path even-odd, and then stroke path
	op_n   = "n\n"  // end path without filling or stroking
)

// clipping path ops (Table 60)
const (
	op_W   = "W\n"  // intersect the clipping path with the current path using the non-zero winding rule
	op_W_X = "W*\n" // intersect the clipping path with the current path using the even-odd rule
)

// color ops (Table 73)
const (
	op_CS  = "CS\n"  // stroking color space
	op_cs  = "cs\n"  // nonstroking color space
	op_SC  = "SC\n"  // stroking color
	op_SCN = "SCN\n" // stroking color with support for additional color spaces
	op_sc  = "sc\n"  // nonstroking color
	op_scn = "scn\n" // nonstroking color with support for additional color spaces
	op_G   = "G\n"   // set stroking color to a DeviceGray color
	op_g   = "g\n"   // set nonstroking color to a DeviceGray color
	op_RG  = "RG\n"  // set stroking color to a DeviceRGB color
	op_rg  = "rg\n"  // set nonstroking color to a DeviceRGB color
	op_K   = "K\n"   // set stroking color to a DeviceCMYK color
	op_k   = "k\n"   // set nonstroking color to a DeviceCMYK color
)

// XObject op (Table 86)
const (
	op_Do = "Do\n" // print XObject
)

// text state ops (Table 103)
const (
	op_Tc = "Tc\n" // character spacing
	op_Tw = "Tw\n" // word spacing
	op_Tz = "Tz\n" // horizontal scaling
	op_TL = "TL\n" // text leading
	op_Tf = "Tf\n" // font size
	op_Tr = "Tr\n" // text rendering mode
	op_Ts = "Ts\n" // text rise
)

// text object ops (Table 105)
const (
	op_BT = "BT\n" // begin text object
	op_ET = "ET\n" // end text object
)

// text positioning ops (Table 106)
const (
	op_Td  = "Td\n" // new line with offset
	op_TD  = "TD\n" // new line with offset and set leading
	op_Tm  = "Tm\n" // text matrix
	op_T_X = "T*\n" // new line and set leading
)

// text showing ops (Table 107)
const (
	op_Tj         = "Tj\n" // show text
	op_APOSTROPHE = "'\n"  // move to new line and show text
	op_QUOTE      = "\"\n" // move to new line and show text with character and word spacing
	op_TJ         = "TJ\n" // show text arrays with glyph positioning
)

// marked content operators (Table 352)
const (
	op_MP  = "MP\n"  // tag marked-content point
	op_DP  = "DP\n"  // tag marked-content point with properties list
	op_BMC = "BMC\n" // begin marked-content sequence
	op_EMC = "EMC\n" // end marked-content sequence
)
