package utils

import (
	"bytes"
	"encoding/base64"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
)

func DecodeBase64(encodedString string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}

func DecodeText(encodedBytes []byte) string {
	decoders := []transform.Transformer{
		unicode.UTF8.NewDecoder(),
		charmap.Windows1252.NewDecoder(),
		japanese.ShiftJIS.NewDecoder(),
		korean.EUCKR.NewDecoder(),
		simplifiedchinese.GB18030.NewDecoder(),
	}

	for _, decoder := range decoders {
		decodedBytes, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(encodedBytes), decoder))
		if err == nil {
			return string(decodedBytes)
		}
	}

	return string(encodedBytes)
}

func ProcessValue(val *interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}
	switch v := (*val).(type) {
	case float64:
		return v, nil
	case []byte:
		strVal := string(v)
		decodedVal, err := base64.StdEncoding.DecodeString(strVal)
		if err != nil {
			// Si la decodificación falla, retorna el string de bytes originales
			return string(v), nil
		}
		return decodedVal, nil
	case nil:
		return nil, nil
	default:
		return *val, nil
	}
}
