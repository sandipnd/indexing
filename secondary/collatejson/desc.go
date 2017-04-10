package collatejson

import "strconv"

//ReverseCollate reverses the bits in an encoded byte stream based
// on the fields specified as desc. Calling reverse on an already
//reversed stream gives back the original stream.
func (codec *Codec) ReverseCollate(code []byte, desc []bool) []byte {

	for i, d := range desc {
		if d {
			field, _, _ := codec.extractEncodedField(code, i+1)
			flipBits(field)
		}
	}
	return code
}

//flips the bits of the given byte slice
func flipBits(code []byte) {

	for i, b := range code {
		code[i] = ^b
	}
	return

}

// get the encoded datum based on Terminator and return a
// tuple of, `encoded-datum`, `remaining-code`, where remaining-code starts
// after the Terminator
func getEncodedDatum(code []byte) (datum []byte, remaining []byte) {
	var i int
	var b byte
	for i, b = range code {
		if b == Terminator || b == ^Terminator {
			break
		}
	}
	return code[:i+1], code[i+1:]
}

func getEncodedString(code []byte) ([]byte, []byte, error) {
	for i := 0; i < len(code); i++ {
		x := code[i]
		if x == Terminator || x == ^Terminator {
			i++
			switch x = code[i]; x {
			case Terminator, ^Terminator:
				if i == (len(code)) {
					return nil, nil, nil
				}
				return code[:i+1], code[i+1:], nil
			default:
				return nil, nil, ErrorSuffixDecoding
			}
			continue
		}
	}
	return nil, nil, ErrorSuffixDecoding
}

//extracts a given field from the encoded byte stream
func (codec *Codec) extractEncodedField(code []byte, fieldPos int) ([]byte, []byte, error) {
	if len(code) == 0 {
		return code, code, nil
	}

	var ts, remaining, datum, orig []byte
	var err error

	orig = code

	switch code[0] {
	case Terminator:
		remaining = code

	case TypeMissing, ^TypeMissing:
		datum, remaining = getEncodedDatum(code)

	case TypeNull, ^TypeNull:
		datum, remaining = getEncodedDatum(code)

	case TypeTrue, ^TypeTrue:
		datum, remaining = getEncodedDatum(code)

	case TypeFalse, ^TypeFalse:
		datum, remaining = getEncodedDatum(code)

	case TypeLength, ^TypeLength:
		datum, remaining = getEncodedDatum(code)

	case TypeNumber, ^TypeNumber:
		datum, remaining = getEncodedDatum(code)

	case TypeString, ^TypeString:
		datum, remaining, err = getEncodedString(code)

	case TypeArray, ^TypeArray:
		var l, currField, currFieldStart int
		if codec.arrayLenPrefix {
			tmp := bufPool.Get().(*[]byte)
			datum, code = getEncodedDatum(code[1:])

			if datum[0] == ^TypeLength {
				flipBits(datum[1:])
			}

			_, ts := DecodeInt(datum[1:len(datum)-1], (*tmp)[:0])
			l, err = strconv.Atoi(string(ts))
			bufPool.Put(tmp)

			if datum[0] == ^TypeLength {
				flipBits(datum[1:])
			}

			currFieldStart = 1
			if err == nil {
				currField = 1
				for ; l > 0; l-- {
					ts, code, err = codec.extractEncodedField(code, 0)
					if err != nil {
						break
					}
					if currField == fieldPos {
						return orig[currFieldStart : currFieldStart+len(ts)], nil, nil
					}
					currField++
					currFieldStart = currFieldStart + len(ts)
				}
			}
		} else {
			code = code[1:]
			currField = 1
			currFieldStart = 1
			for code[0] != Terminator && code[0] != ^Terminator {
				ts, code, err = codec.extractEncodedField(code, 0)
				if err != nil {
					break
				}
				if currField == fieldPos {
					return orig[currFieldStart : currFieldStart+len(ts)], nil, nil
				}
				currField++
				currFieldStart = currFieldStart + len(ts)
				datum = append(datum, ts...)

			}
		}
		remaining = code[1:]             // remove Terminator
		datum = orig[0 : len(datum)+1+1] //1 for Terminator and 1 for Type

	case TypeObj, ^TypeObj:
		var l int
		var key, value []byte
		if codec.propertyLenPrefix {
			tmp := bufPool.Get().(*[]byte)
			datum, code = getEncodedDatum(code[1:])

			if datum[0] == ^TypeLength {
				flipBits(datum[1:])
			}

			_, ts := DecodeInt(datum[1:len(datum)-1], (*tmp)[:0])
			l, err = strconv.Atoi(string(ts))
			bufPool.Put(tmp)

			if datum[0] == ^TypeLength {
				flipBits(datum[1:])
			}

			if err == nil {
				for ; l > 0; l-- {
					key, code, err = codec.extractEncodedField(code, 0)
					if err != nil {
						break
					}
					value, code, err = codec.extractEncodedField(code, 0)
					if err != nil {
						break
					}
					datum = append(datum, key...)
					datum = append(datum, value...)
				}
			}
		} else {
			code = code[1:]
			for code[0] != Terminator && code[0] != ^Terminator {
				key, code, err = codec.extractEncodedField(code, 0)
				if err != nil {
					break
				}
				value, code, err = codec.extractEncodedField(code, 0)
				if err != nil {
					break
				}
				datum = append(datum, key...)
				datum = append(datum, value...)
			}
		}
		remaining = code[1:]             // remove Terminator
		datum = orig[0 : len(datum)+1+1] //1 for Terminator and 1 for Type
	}
	return datum, remaining, err
}