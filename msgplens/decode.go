package main

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func decodePy3Bytes(input []byte) ([]byte, error) {
	// b'\t\x00\x00\x00\t\x00\x00\x00\x03\x00\x00\x00\x00\x00\x00\x00\x91\xa7content'
	out := make([]byte, 0, len(input))

	cur := 0
	quoted := false
	if input[cur] == 'b' {
		cur++
	}
	if input[cur] == '\'' {
		quoted = true
		cur++
	}

	stateString, stateEscOpen, stateEscHex := 0, 1, 2
	state := stateString

	end := len(input)
	done := false
	i := cur
	for ; i < end && !done; i++ {
		switch state {
		case stateString:
			if input[i] == '\\' {
				state = stateEscOpen
			} else if quoted && input[i] == '\'' {
				done = true
			} else {
				out = append(out, input[i])
			}

		case stateEscOpen:
			if input[i] == 'x' {
				state = stateEscHex
			} else {
				b := byte(0)
				switch input[i] {
				case '\n':
					// do nothing
				case '\\':
					b = '\\'
				case '\'':
					b = '\''
				case '"':
					b = '"'
				case 'a':
					b = 7
				case 'b':
					b = 8
				case 'f':
					b = 12
				case 'n':
					b = '\n'
				case 'r':
					b = '\r'
				case 't':
					b = '\t'
				case 'v':
					b = 11
				default:
					b = input[i]
					out = append(out, '\\')
				}
				if b != 0 {
					out = append(out, b)
				}
				state = stateString
			}

		case stateEscHex:
			if end-i < 2 {
				return nil, errors.Errorf("incomplete hex")
			}
			c, err := strconv.ParseInt(string(input[i:i+2]), 16, 0)
			if err != nil {
				return nil, err
			}
			out = append(out, byte(c))
			state = stateString
			i++
		}
	}

	for i < end && (input[i] == '\n' || input[i] == '\r' || input[i] == '\t') {
		i++
	}

	if i != end {
		return nil, errors.Errorf("did not read to end")
	}

	return out, nil
}

var numSplit = regexp.MustCompile("[^0-9a-fA-Fx]+")

func decodeNums(input []byte, base int) ([]byte, error) {
	input = bytes.Trim(input, "[] ")
	parts := numSplit.Split(string(input), -1)

	out := make([]byte, 0, len(input))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if len(p) == 0 {
			continue
		}
		i, err := strconv.ParseUint(p, base, 8)
		if err != nil {
			return nil, err
		}
		out = append(out, byte(i))
	}
	return out, nil
}
