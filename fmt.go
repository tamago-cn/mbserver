package mbserver

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// FormatBytes formate bytes to string
func FormatBytes(in []byte) string {
	n := len(in)
	bs := make([]string, n)
	for i := 0; i < n; i++ {
		bs[i] = fmt.Sprintf("%02X", in[i])
	}
	return strings.Join(bs, " ")
}

// HexStrToBytes parse HEX string to bytes
func HexStrToBytes(in string) ([]byte, error) {
	in = strings.ReplaceAll(in, " ", "")
	n := len(in)
	if n%2 != 0 {
		return []byte{}, errors.New("hexStrToBytes error: input not valid hex string")
	}
	bs := make([]byte, n/2)
	for i := 0; i < n; i += 2 {
		b, err := strconv.ParseInt(in[i:i+2], 16, 16)
		if err != nil {
			return []byte{}, err
		}
		bs[i/2] = byte(b)
	}
	return bs, nil
}
