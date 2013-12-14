package properties

import (
	"fmt"
	"io"
	"io/ioutil"
)

type Decoder struct {
	r io.Reader
}

type Encoding uint

const (
	UTF8 Encoding = 1 << iota
	ISO_8859_1
)

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode() (*Properties, error) {
	return decode(d.r, ISO_8859_1)
}

func (d *Decoder) DecodeWithEncoding(enc Encoding) (*Properties, error) {
	return decode(d.r, enc)
}

func decode(r io.Reader, enc Encoding) (*Properties, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return newParser().Parse(convert(buf, enc))
}

// The Java properties spec says that .properties files must be ISO-8859-1
// encoded. Since the first 256 unicode code points cover ISO-8859-1 we
// can convert each byte into a rune and use the resulting string
// as UTF-8 input for the parser.
func convert(buf []byte, enc Encoding) string {
	switch enc {
	case UTF8:
		return string(buf)
	case ISO_8859_1:
		runes := make([]rune, len(buf))
		for i, b := range buf {
			runes[i] = rune(b)
		}
		return string(runes)
	default:
		panic(fmt.Sprintf("unsupported encoding %v", enc))
	}
}
