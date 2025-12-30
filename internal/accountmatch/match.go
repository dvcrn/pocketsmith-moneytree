package accountmatch

import (
	"fmt"
	"strings"

	"github.com/dvcrn/pocketsmith-go"
)

func normalizeAccountTitle(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(title), " "))
}

func BuildDisplayAccountName(institutionName, baseName string) string {
	if strings.TrimSpace(institutionName) == "" {
		return baseName
	}

	return fmt.Sprintf("%s - %s", institutionName, baseName)
}

func FindMatchingAccount(accounts []*pocketsmith.Account, institutionName, baseName, displayName string) (*pocketsmith.Account, error) {
	normalizedBase := normalizeAccountTitle(baseName)
	normalizedDisplay := normalizeAccountTitle(displayName)
	normalizedInstitution := normalizeAccountTitle(institutionName)

	var displayMatches []*pocketsmith.Account
	var baseMatches []*pocketsmith.Account
	var suffixMatches []*pocketsmith.Account
	var institutionMatches []*pocketsmith.Account

	for _, account := range accounts {
		normalizedTitle := normalizeAccountTitle(account.Title)
		if normalizedTitle == normalizedDisplay {
			displayMatches = append(displayMatches, account)
			continue
		}
		if normalizedTitle == normalizedBase {
			baseMatches = append(baseMatches, account)
			continue
		}

		if strings.HasSuffix(normalizedTitle, normalizedBase) {
			suffixMatches = append(suffixMatches, account)

			if normalizedInstitution != "" {
				accountInstitution := normalizeAccountTitle(account.PrimaryTransactionAccount.Institution.Title)
				if accountInstitution == normalizedInstitution {
					institutionMatches = append(institutionMatches, account)
				}
			}
		}
	}

	if len(displayMatches) == 1 {
		return displayMatches[0], nil
	}
	if len(displayMatches) > 1 {
		return nil, fmt.Errorf("multiple Pocketsmith accounts match %q", displayName)
	}
	if len(baseMatches) == 1 {
		return baseMatches[0], nil
	}
	if len(baseMatches) > 1 {
		return nil, fmt.Errorf("multiple Pocketsmith accounts match %q", baseName)
	}
	if len(institutionMatches) == 1 {
		return institutionMatches[0], nil
	}
	if len(institutionMatches) > 1 {
		return nil, fmt.Errorf("multiple Pocketsmith accounts match %q for institution %q", baseName, institutionName)
	}
	if len(suffixMatches) == 1 {
		return suffixMatches[0], nil
	}
	if len(suffixMatches) > 1 {
		return nil, fmt.Errorf("multiple Pocketsmith accounts match %q; rename to disambiguate", baseName)
	}

	return nil, pocketsmith.ErrNotFound
}
