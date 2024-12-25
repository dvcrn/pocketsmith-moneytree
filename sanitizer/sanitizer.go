package sanitizier

import (
	"strings"

	"github.com/dvcrn/romajiconv"
)

func Sanitize(input string) string {
	convertedPayee := romajiconv.ConvertFullWidthToHalf(input)

	// when /iD is present, add a space so that it's " /iD" instead of "/iD"
	convertedPayee = strings.ReplaceAll(convertedPayee, "/iD", " /iD")
	convertedPayee = strings.ReplaceAll(convertedPayee, "/NFC", " /NFC")

	// strip all consecutive spaces
	convertedPayee = strings.Join(strings.Fields(convertedPayee), " ")

	return convertedPayee
}
