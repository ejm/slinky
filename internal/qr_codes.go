package internal

import qrcode "github.com/skip2/go-qrcode"

func createQRCode(url string, size int) (png []byte, err error) {
	if size > 1024 {
		size = 1024
	} else if size < 0 {
		size = 0
	}
	png, err = qrcode.Encode(url, qrcode.Medium, size)
	return
}
