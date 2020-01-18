package encode

import (
	"bytes"
	"encoding/binary"
	"net/url"
	"sort"
)

func BytesCombine(pBytes ...[]byte) []byte {
	len := len(pBytes)
	s := make([][]byte, len)

	for index := 0; index < len; index++ {
		s[index] = pBytes[index]
	}

	sep := []byte("")

	return bytes.Join(s, sep)
}

func EncodeByte(body []byte, v byte) []byte {
	return append(body, v)
}

func DecodeByte(body []byte) ([]byte, byte) {
	return body[1:], body[0]
}

func EncodeUint16(body []byte, v uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, v)

	return BytesCombine(body, buf)
}

func DecodeUint16(body []byte) ([]byte, uint16) {
	return body[2:], binary.BigEndian.Uint16(body)
}

func EncodeUint32(body []byte, v uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)

	return BytesCombine(body, buf)
}

func DecodeUint32(body []byte) ([]byte, uint32) {
	return body[4:], binary.BigEndian.Uint32(body)
}

func UrlEncode(urlParasMap map[string]string) string {
	v := url.Values{}

	var keys []string
	for k := range urlParasMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		if len(urlParasMap[key]) > 0 {
			v.Add(key, urlParasMap[key])
		}
	}

	urlParasString := v.Encode()

	return urlParasString
}
