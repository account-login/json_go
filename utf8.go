package json_go

import "fmt"

type DecodingError struct {
	pos  int
	char byte
	msg  string
}

func (err *DecodingError) Error() string {
	return fmt.Sprintf("DecodingError at %d, byte=%#x, msg=%s",
		err.pos, err.char, err.msg)
}

func ReadCode(buf []byte, cur int) (code rune, next int, err error) {
	if len(buf)-cur <= 0 {
		err = &DecodingError{cur, 0, "no enough data"}
		return
	}

	var mask byte
	leading := buf[cur]
	switch {
	case leading < 0x80: // prefix 0
		next = cur + 1
		mask = 0xff
	case leading < 0xc0: // prefix 10
		err = &DecodingError{cur, leading, "unexpected leading char"}
	case leading < 0xe0: // prefix 110
		next = cur + 2
		mask = 0x1f
	case leading < 0xf0: // prefix 1110
		next = cur + 3
		mask = 0x0f
	case leading < 0xf8: // prefix 11110
		next = cur + 4
		mask = 0x07
	default:
		err = &DecodingError{cur, leading, "bad leading char"}
	}

	if err != nil {
		return
	}

	numFollowing := next - cur - 1
	if numFollowing+1 > len(buf)-cur {
		err = &DecodingError{
			cur, leading,
			fmt.Sprintf("buf not enough. req: %d, remain: %d",
				numFollowing+1, len(buf)-cur)}
		return
	}

	code = rune(leading & mask)
	for i := 0; i < numFollowing; i++ {
		code <<= 6
		code |= rune(buf[cur+1+i] & 0x3f)
	}

	return
}

func Decode(input []byte) (output []rune, err error) {
	for cur := 0; cur < len(input); {
		var code rune
		code, cur, err = ReadCode(input, cur)
		if err != nil {
			return
		}
		output = append(output, code)
	}
	return
}

func DecodeString(input string) (output []rune, err error) {
	return Decode([]byte(input))
}
