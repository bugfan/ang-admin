package service

import (
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// GenerateCaptcha creates a new captcha and returns its id and base64 string
func GenerateCaptcha() (id, b64s string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	cp := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err = cp.Generate()
	return id, b64s, err
}

// VerifyCaptcha verifies the captcha code
func VerifyCaptcha(id, answer string) bool {
	return store.Verify(id, answer, true)
}
