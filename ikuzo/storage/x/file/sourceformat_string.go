// Code generated by "stringer -type=SourceFormat"; DO NOT EDIT.

package file

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[FormatRaw-0]
	_ = x[FormatNTriples-1]
	_ = x[FormatEAD-2]
}

const _SourceFormat_name = "FormatRawFormatNTriplesFormatEAD"

var _SourceFormat_index = [...]uint8{0, 9, 23, 32}

func (i SourceFormat) String() string {
	if i < 0 || i >= SourceFormat(len(_SourceFormat_index)-1) {
		return "SourceFormat(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _SourceFormat_name[_SourceFormat_index[i]:_SourceFormat_index[i+1]]
}