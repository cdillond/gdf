package gdf

const (
	// graphics state ops (Table 56
	op_q  = "q\n"
	op_Q  = "Q\n"
	op_cm = "cm\n"
	op_w  = "w\n"
	op_J  = "J\n"
	op_j  = "j\n"
	op_M  = "M\n"
	op_d  = "d\n"
	op_ri = "ri\n"
	op_i  = "i\n"
	op_gs = "gs\n"

	// path construction ops (Table 58
	op_m  = "m\n"
	op_l  = "l\n"
	op_c  = "c\n"
	op_v  = "v\n"
	op_y  = "y\n"
	op_h  = "h\n"
	op_re = "re\n"

	// path painting ops (Table 59
	op_S   = "S\n"
	op_s   = "s\n"
	op_f   = "f\n"
	op_f_X = "f*\n"
	op_B   = "B\n"
	op_B_X = "B*\n"
	op_b   = "b\n"
	op_b_X = "b*\n"
	op_n   = "n\n"

	// clipping path ops (Table 60
	op_W   = "W\n"
	op_W_X = "W*\n"

	// color ops (Table 73; upper
	op_CS  = "CS\n"
	op_cs  = "cs\n"
	op_SC  = "SC\n"
	op_SCN = "SCN\n"
	op_sc  = "sc\n"
	op_scn = "scn\n"
	op_G   = "G\n"
	op_g   = "g\n"
	op_RG  = "RG\n"
	op_rg  = "rg\n"
	op_K   = "K\n"
	op_k   = "k\n"

	// XObject op (Table 86
	op_Do = "Do\n"

	// text state ops (Table 103
	op_Tc = "Tc\n"
	op_Tw = "Tw\n"
	op_Tz = "Tz\n"
	op_TL = "TL\n"
	op_Tf = "Tf\n"
	op_Tr = "Tr\n"
	op_Ts = "Ts\n"

	// text object ops (Table 105
	op_BT = "BT\n"
	op_ET = "ET\n"

	// text positioning ops (Table 106
	op_Td  = "Td\n"
	op_TD  = "TD\n"
	op_Tm  = "Tm\n"
	op_T_X = "T*\n"

	// text showing ops (Table 107
	op_Tj         = "Tj\n"
	op_APOSTROPHE = "'\n"
	op_QUOTE      = "\"\n"
	op_TJ         = "TJ\n"

	// marked content operators (Table 352
	op_MP  = "MP\n"
	op_DP  = "DP\n"
	op_BMC = "BMC\n"
	op_EMC = "EMC\n"
)
