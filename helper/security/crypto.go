package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/derekyu332/goii/helper/extend"
)

func MD5(input string) string {
	return fmt.Sprintf("%x", md5.Sum(extend.StringBytes(input)))
}

func SHA1(input string) string {
	return fmt.Sprintf("%x", sha1.Sum(extend.StringBytes(input)))
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])

	if unpadding < length {
		return origData[:(length - unpadding)]
	} else {
		return origData
	}
}

func BASE64Encode(input string) string {
	var data []byte = extend.StringBytes(input)
	encodeString := base64.StdEncoding.EncodeToString(data)

	return encodeString
}

func HMACSHA1Encode(accessSecret string, stringToSign string) string {
	key := extend.StringBytes(accessSecret)
	mac := hmac.New(sha1.New, key)
	data := extend.StringBytes(stringToSign)
	_, err := mac.Write(data)

	if err != nil {
		return ""
	}

	return string(mac.Sum(nil)[:])
}

func HMACSHA1EncodeHex(accessSecret string, stringToSign string) string {
	key := extend.StringBytes(accessSecret)
	mac := hmac.New(sha1.New, key)
	data := extend.StringBytes(stringToSign)
	_, err := mac.Write(data)

	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", mac.Sum(nil)[:])
}
