package json_go

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Good(t *testing.T, input string, expect JsonValue) {
	got, err := Parse(input)
	if assert.NoError(t, err) {
		assert.Equal(t, expect, got)
	}
	got, err = Parse(" " + input + " ")
	if assert.NoError(t, err) {
		assert.Equal(t, expect, got)
	}
}

func Bad(t *testing.T, input string) {
	_, err := Parse(input)
	assert.Error(t, err)
	t.Log(input, "\t", err)
}

func TestParseNum(t *testing.T) {
	goodi := func(input string, expect int64) { Good(t, input, expect) }
	goodf := func(input string, expect float64) { Good(t, input, expect) }
	bad := func(input string) { Bad(t, input) }

	bad("")

	goodi("123", 123)
	goodi("0", 0)
	bad("00")
	goodi("-0", 0)
	bad("-00")
	goodi("-124", -124)

	goodf("1.2", 1.2)
	goodf("-1.2", -1.2)
	goodf("0.2", 0.2)
	goodf("-0.2", -0.2)
	goodf("-0.02", -0.02)
	goodf("12.2", 12.2)

	bad("1.")
	bad(".1")
	bad("-.1")

	goodf("1e2", 100.0)
	goodf("1e+2", 100.0)
	goodf("1e-2", 0.01)
	goodf("1e0", 1.0)

	bad("1e")
	bad("1e+")
	bad("1e-")
	bad("1e.")

	bad("1.e1")
}

func TestParseArray(t *testing.T) {
	good := func(input string, expect ...JsonValue) {
		if expect == nil {
			expect = []JsonValue{}
		}
		Good(t, input, JsonArray(expect))
	}
	bad := func(input string) { Bad(t, input) }

	good("[]")
	good("[ ]")
	bad("[")
	bad("[,]")

	good("[123]", int64(123))
	good("[123, -1e-1]", int64(123), -0.1)
	good("[123,-1e-1]", int64(123), -0.1)
	good("[ 123, -1e-1 ]", int64(123), -0.1)

	bad("[124,]")
	bad("[124 124]")
}

func TestParseBoolNull(t *testing.T) {
	good := func(input string, expect JsonValue) { Good(t, input, expect) }
	bad := func(input string) { Bad(t, input) }

	good("null", nil)
	good("true", true)
	good("false", false)

	bad("nul")
	bad("nulll")
}

func TestParseString(t *testing.T) {
	good := func(input string, expect JsonValue) { Good(t, input, expect) }
	bad := func(input string) { Bad(t, input) }

	good(`""`, "")
	bad(`"`)
	good(`"\"\\\/\b\f\n\r\t"`, "\"\\/\b\f\n\r\t")
	bad(`"\"`)
	bad(`"\`)
	good(`"\u1234"`, "\u1234")
	good(`"\uaaaa"`, "\uaaaa")
	good(`"\uAAAA"`, "\uaaaa")
	bad(`"\u124"`)
	bad(`"\u124`)
	bad(`"\u124g"`)

	bad(`"\x"`)
	bad("\"\x19\"")
}

func TestParseMap(t *testing.T) {
	good := func(input string, expect JsonValue) { Good(t, input, expect) }
	bad := func(input string) { Bad(t, input) }

	good("{}", JsonMap{})
	bad("{")
	good(`{"a": 123}`, JsonMap{"a": int64(123)})
	good(`{"b": 23, "c": "d"}`, JsonMap{"b": int64(23), "c": "d"})
	bad(`{"b": 23, "c": "d",}`)
	bad(`{"b": ,}`)
	bad(`{"b": }`)
	bad(`{"b", "c": 1}`)
}
