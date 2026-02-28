package internal

import (
	"fmt"
	"testing"
)

type encodeTest struct {
	inp  []byte
	want string
}

func TestEncode(t *testing.T) {
	tests := []encodeTest{
		{[]byte{0x0f}, "bh"},
		{[]byte{0xab, 0xcd}, "qxtw"},
		{[]byte{0x00, 0x01, 0x02}, "bbbnbd"},
		{[]byte{0x12, 0x34, 0x56}, "ndrfgk"},
		{[]byte{0xe6, 0x7e, 0x22}, "zkmzdd"},
		{[]byte{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, "bndrfgkmcpqxtwzh"},
	}

	for _, test := range tests {
		name := fmt.Sprintf("encode(%v)", test.inp)
		t.Run(name, func(tt *testing.T) {
			got := encode(test.inp)
			if test.want != got {
				t.Errorf("encode(%v) = %s; want %s", test.inp, got, test.want)
			}
		})
	}
}
