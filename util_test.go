package gomocker

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestParseInt(t *testing.T) {
	testcases := []struct {
		name     string
		number   string
		expected int
		isValid  bool
	}{
		{name: "42", number: "42", expected: 42, isValid: true},
		{name: "4_2", number: "4_2", expected: 4_2, isValid: true},
		{name: "0600", number: "0600", expected: 0600, isValid: true},
		{name: "0_600", number: "0_600", expected: 0_600, isValid: true},
		{name: "0o600", number: "0o600", expected: 0o600, isValid: true},
		{name: "0O600", number: "0O600", expected: 0O600, isValid: true},
		{name: "0xBadFace", number: "0xBadFace", expected: 0xBadFace, isValid: true},
		{name: "0xBad_Face", number: "0xBad_Face", expected: 0xBad_Face, isValid: true},
		{name: "0x_67_7a_2f_cc_40_c6", number: "0x_67_7a_2f_cc_40_c6", expected: 0x_67_7a_2f_cc_40_c6, isValid: true},
		{name: "17011692133845727", number: "17011692133845727", expected: 17011692133845727, isValid: true},
		{name: "170_113_49_3_633_754_127", number: "170_113_49_3_633_754_127", expected: 170_113_49_3_633_754_127, isValid: true},
		{name: "_42", number: "_42", expected: 0, isValid: false},
		{name: "42_", number: "42_", expected: 0, isValid: false},
		{name: "4__2", number: "4__2", expected: 0, isValid: false},
		{name: "0_xBadFace", number: "0_xBadFace", expected: 0, isValid: false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			v, ok := parseInt(tc.number)
			assert.Equal(t, v, tc.expected)
			assert.Equal(t, ok, tc.isValid)
		})
	}
}
