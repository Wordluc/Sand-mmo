package wasm

import (
	"strings"
	"syscall/js"
)

var SIZE_CELL = 4

func IsMobile() bool {
	userAgent := js.Global().Get("navigator").Get("userAgent").String()
	mobileKeywords := []string{"Android", "iPhone", "Mobile"}
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}
	return false
}
func init() {
	if IsMobile() {
		SIZE_CELL = 2
	}
}
