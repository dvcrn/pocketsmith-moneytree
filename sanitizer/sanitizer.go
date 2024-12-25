package sanitizier

import (
	"github.com/dvcrn/romajiconv"
	"strings"
)

func Sanitize(input string) string {
	convertedPayee := romajiconv.ConvertFullWidthToHalf(input)

	// when /iD is present, add a space so that it's " /iD" instead of "/iD"
	convertedPayee = strings.ReplaceAll(convertedPayee, "/iD", " /iD")

	// strip all consecutive spaces
	convertedPayee = strings.Join(strings.Fields(convertedPayee), " ")

	return convertedPayee
}
