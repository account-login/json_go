package json_go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadCode(t *testing.T) {
	const pad = "pad!"
	bad := func(input string) {
		_, _, err := ReadCode([]byte(pad+input), len(pad))
		t.Log([]byte(input), err)
		assert.Error(t, err)
	}

	good := func(input string, expectCode int, size int) {
		code, next, err := ReadCode([]byte(pad+input), len(pad))
		if assert.NoError(t, err) {
			assert.Equal(t, rune(expectCode), code)
			assert.Equal(t, size+len(pad), next)
		}
	}

	bad("")
	bad("\xff")
	bad("啊"[1:])
	bad("啊"[:1])
	bad("啊"[:2])

	good("a", 'a', 1)
	good("啊", 0x554a, 3)
	good("\xf4\x8f\xbf\xbf", 0x10ffff, 4)
}

func TestDecodeString(t *testing.T) {
	good := func(input string) {
		output, err := DecodeString(input)
		if assert.NoError(t, err) {
			assert.Equal(t, []rune(input), output)
		}
	}

	bad := func(input string) {
		part, err := DecodeString(input)
		if assert.Error(t, err) {
			t.Log(input, part, err)
		}
	}

	good("asdf啊124")
	bad("asdf啊\xfe124")
}
