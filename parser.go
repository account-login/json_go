package json_go

import (
	"fmt"
	"math"
)

type JsonValue interface{} // float64, int64, bool, nil, JsonMap, JsonArray
type JsonMap map[string]JsonValue
type JsonArray []JsonValue

type JsonKeyValue struct {
	key   string
	value JsonValue
}

func SkipSpace(input []rune, cur int) int {
	for i := cur; i < len(input); i++ {
		switch input[i] {
		case ' ', '\t', '\n', '\r':
		default:
			return i
		}
	}
	return len(input)
}

func Consume(input []rune, cur int, tok string) (next int, err error) {
	next = SkipSpace(input, cur)
	tokrune := []rune(tok)
	if len(input)-next < len(tokrune) {
		err = &ParseError{next, fmt.Sprintf("expect %q", tok)}
		return
	}

	for i, ch := range tokrune {
		if input[next+i] != ch {
			err = &ParseError{next + i, fmt.Sprintf("expect %q", tok)}
			return
		}
	}
	next += len(tokrune)
	return
}

type ParseError struct {
	pos int
	msg string
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("ParseError at %d: %s", err.pos, err.msg)
}

func Parse(input string) (value JsonValue, err error) {
	var decoded []rune
	decoded, err = DecodeString(input)
	if err != nil {
		return
	}
	return ParseRunes(decoded)
}

func ParseRunes(input []rune) (value JsonValue, err error) {
	var next int
	value, next, err = ParseAny(input, 0)

	if err == nil {
		next = SkipSpace(input, next)
		if next != len(input) {
			err = &ParseError{next, "not terminated"}
		}
	}
	return
}

func ParseAny(input []rune, cur int) (value JsonValue, next int, err error) {
	next = SkipSpace(input, cur)
	if next >= len(input) {
		err = &ParseError{next, "expect something, got EOS"}
		return
	}

	switch input[next] {
	case '[':
		value, next, err = ParseArray(input, next)
	case '{':
		value, next, err = ParseMap(input, next)
	case '"':
		value, next, err = ParseString(input, next)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		value, next, err = ParseNum(input, next)
	case 't', 'f', 'n':
		value, next, err = ParseBoolNull(input, next)
	default:
		err = &ParseError{next, fmt.Sprintf("bad char: '%c' (%#x)", input[next], input[next])}
	}

	return
}

func IsNoEscape(ch rune) bool {
	return (0x23 <= ch && ch <= 0x5b) || (0x5d <= ch && ch <= 0x10ffff) || ch == ' ' || ch == '!'
}

func Hex2Num(input []rune, cur int) (value rune, err error) {
	ch := input[cur]
	switch {
	case '0' <= ch && ch <= '9':
		value = ch - '0'
	case 'a' <= ch && ch <= 'f':
		value = ch - 'a' + 10
	case 'A' <= ch && ch <= 'F':
		value = ch - 'A' + 10
	default:
		err = &ParseError{cur, fmt.Sprintf("expect hex, got '%c' (%#x)", ch, ch)}
	}
	return
}

func ScanHex(input []rune, cur int) (value rune, err error) {
	if cur+4 > len(input) {
		err = &ParseError{cur, "expect 4 hex digit"}
		return
	}

	for i := 0; i < 4; i++ {
		var d rune
		d, err = Hex2Num(input, cur+i)
		if err != nil {
			return
		}

		value <<= 4
		value |= d
	}

	return
}

func ParseEscape(input []rune, cur int) (value rune, next int, err error) {
	next = cur
	if cur >= len(input) {
		err = &ParseError{next, "string not terminated, expect escape"}
		return
	}

	ch := input[next]
	switch ch {
	case '"', '\\', '/':
		value = ch
	case 'b':
		value = '\b'
	case 'f':
		value = '\f'
	case 'n':
		value = '\n'
	case 'r':
		value = '\r'
	case 't':
		value = '\t'
	case 'u':
		next++
		value, err = ScanHex(input, next)
		if err != nil {
			return
		}
		next += 3
	default:
		err = &ParseError{next, fmt.Sprintf("bad escape char: '%c' (%#x)", ch, ch)}
		return
	}
	next++

	return
}

func ParseString(input []rune, cur int) (value string, next int, err error) {
	next, err = Consume(input, cur, "\"")
	if err != nil {
		return
	}

	val := []rune{}
	for next < len(input) {
		ch := input[next]
		switch {
		case ch == '"': // terminated
			value = string(val)
			next++
			return
		case ch == '\\':
			next++
			ch, next, err = ParseEscape(input, next)
			if err != nil {
				return
			}
			val = append(val, ch)
		case IsNoEscape(ch):
			val = append(val, ch)
			next++
		default:
			err = &ParseError{next, fmt.Sprintf("unescaped char: '%c' (%#x)", ch, ch)}
			return
		}
	}

	err = &ParseError{next, "string not terminated"}
	return
}

func IsDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func ScanInt(input []rune, cur int) (value int64, next int, err error) {
	if !(cur < len(input) && IsDigit(input[cur])) {
		err = &ParseError{cur, "expect digits"}
		return
	}

	for next = cur; next < len(input) && IsDigit(input[next]); next++ {
		value *= 10
		value += int64(input[next] - '0')
	}
	return
}

func ParseNum(input []rune, cur int) (value JsonValue, next int, err error) {
	neg := false
	var suberr error
	next, suberr = Consume(input, cur, "-")
	if suberr == nil {
		neg = true
	}

	// unreachable
	if next >= len(input) {
		err = &ParseError{next, "expects digits, got EOS"}
		return
	}

	// integer part
	var whole int64
	if input[next] == '0' {
		whole = 0
		next++
	} else {
		whole, next, err = ScanInt(input, next)
		if err != nil {
			return
		}
	}

	// frac part
	isfloat := false
	frac := float64(0)
	if next < len(input) && input[next] == '.' {
		isfloat = true
		next++

		if !(next < len(input) && IsDigit(input[next])) {
			err = &ParseError{next, "expect digits"}
			return
		}

		scale := float64(10)
		for ; next < len(input) && IsDigit(input[next]); next++ {
			frac += float64(input[next]-'0') / scale
			scale *= 10
		}
	}

	// exp part
	hasexp := false
	expnum := int64(0)
	if next < len(input) && (input[next] == 'e' || input[next] == 'E') {
		next++
		isfloat = true
		hasexp = true
		expneg := false

		for sign, val := range map[string]bool{"+": false, "-": true} {
			next, err = Consume(input, next, sign)
			if err == nil {
				expneg = val
				break
			}
		}

		expnum, next, err = ScanInt(input, next)
		if err != nil {
			return
		}
		if expneg {
			expnum = -expnum
		}
	}

	if !isfloat {
		if neg {
			value = -whole
		} else {
			value = whole
		}
	} else {
		fval := float64(whole)
		fval += frac
		if hasexp {
			fval *= math.Pow10(int(expnum))
		}

		if neg {
			fval = -fval
		}
		value = fval
	}

	return
}

func ParseBoolNull(input []rune, cur int) (value JsonValue, next int, err error) {
	var suberr error
	for literal, val := range map[string]JsonValue{"true": true, "false": false, "null": nil} {
		next, suberr = Consume(input, cur, literal)
		if suberr == nil {
			value = val
			return
		}
	}

	err = &ParseError{next, "expect true|false|null"}
	return
}

func ParseMap(input []rune, cur int) (value JsonValue, next int, err error) {
	value, next, err = ParseArrayLike(input, cur, ParseKeyValue, [2]string{"{", "}"})

	// convert array to map
	if err == nil {
		jmap := JsonMap{}
		for _, item := range value.(JsonArray) {
			kv := item.(JsonKeyValue)
			// TODO: warn duplicated key
			jmap[kv.key] = kv.value
		}
		value = jmap
	}
	return
}

func ParseKeyValue(input []rune, cur int) (value JsonValue, next int, err error) {
	var kv JsonKeyValue
	kv.key, next, err = ParseString(input, cur)
	if err != nil {
		return
	}

	next, err = Consume(input, next, ":")
	if err != nil {
		return
	}

	kv.value, next, err = ParseAny(input, next)
	if err != nil {
		return
	}

	value = kv
	return
}

func ParseArray(input []rune, cur int) (value JsonValue, next int, err error) {
	return ParseArrayLike(input, cur, ParseAny, [2]string{"[", "]"})
}

type ParseFunc func(input []rune, cur int) (value JsonValue, next int, err error)

func ParseArrayLike(input []rune, cur int, itemParser ParseFunc, bracket [2]string) (value JsonValue, next int, err error) {
	next, err = Consume(input, cur, bracket[0])
	if err != nil { // unreachable
		return
	}

	// empty array []
	var suberr error
	next, suberr = Consume(input, next, bracket[1])
	if suberr == nil {
		value = JsonArray{}
		return
	}

	var subval JsonValue
	arr := JsonArray{}
	for {
		subval, next, err = itemParser(input, next)
		if err != nil {
			return
		}
		arr = append(arr, subval)

		next, suberr = Consume(input, next, ",")
		if suberr == nil {
			continue
		}
		next, suberr = Consume(input, next, bracket[1])
		if suberr != nil {
			err = &ParseError{next, fmt.Sprintf("expect '%s' or ','", bracket[1])}
			return
		} else {
			break
		}
	}

	value = arr
	return
}
