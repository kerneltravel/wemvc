package wemvc

import "regexp"

const (
	Version = "0.1"
)

var regNumber *regexp.Regexp
var regString *regexp.Regexp
var regRouteKey *regexp.Regexp

func init() {
	regNumber, _ = regexp.Compile(`^\d+`)
	regString, _ = regexp.Compile(`^\w+`)
	regRouteKey, _ = regexp.Compile(`{\w+}`)
}
