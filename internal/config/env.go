package config

import (
	"fmt"
	"os"
)

var ADDR = fmt.Sprintf("%s:%s", "0.0.0.0", "9999")
var DefaultURL = GetDefaultEnv("DEFAULT_PP", "http://payment-processor-default:8080")
var FallbackURL = GetDefaultEnv("FALLBACK_PP", "http://payment-processor-fallback:8080")
var Token = GetDefaultEnv("TOKEN", "123")



func GetDefaultEnv(k, d string) string {
	if a, exist := os.LookupEnv(k); exist {
		return a
	}
	return d
}
