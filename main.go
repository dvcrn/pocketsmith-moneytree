package main

import (
	"flag"
	"fmt"
	sanitizier "github.com/dvcrn/pocketsmith-anapay/sanitizer"
	"os"
	"strings"
	"time"

	"github.com/dvcrn/pocketsmith-anapay/moneytree"
	"github.com/dvcrn/pocketsmith-go"
)

type Config struct {
	MoneytreeUsername string
	MoneytreePassword string
	MoneytreeApiKey   string
	PocketsmithToken  string

	NumTransactions int
}

func getConfig() *Config {
	config := &Config{}

	// Define command-line flags
	flag.StringVar(&config.MoneytreeUsername, "username", os.Getenv("MONEYTREE_USERNAME"), "Moneytree username")
	flag.StringVar(&config.MoneytreePassword, "password", os.Getenv("MONEYTREE_PASSWORD"), "Moneytree password")
	flag.StringVar(&config.MoneytreeApiKey, "apikey", os.Getenv("MONEYTREE_API_KEY"), "Moneytree API KEY")

	flag.StringVar(&config.PocketsmithToken, "pocketsmith-token", os.Getenv("POCKETSMITH_TOKEN"), "Pocketsmith API token")
	flag.Parse()

	// Validate required fields
	if config.MoneytreeUsername == "" {
		fmt.Println("Error: Moneytree username is required. Set via -username flag or MONEYTREE_USERNAME environment variable")
		os.Exit(1)
	}
	if config.MoneytreePassword == "" {
		fmt.Println("Error: Moneytree password is required. Set via -password flag or MONEYTREE_PASSWORD environment variable")
		os.Exit(1)
	}
	if config.MoneytreeApiKey == "" {
		fmt.Println("Error: Moneytree API KEY is required. Set via -apikey flag or MONEYTREE_API_KEY environment variable")
		os.Exit(1)
	}
	if config.PocketsmithToken == "" {
		fmt.Println("Error: Pocketsmith token is required. Set via -token flag or POCKETSMITH_TOKEN environment variable")
		os.Exit(1)
	}

	return config
}

func findCredentialFromMeta(gm *moneytree.MTGuest, credentialID int) *moneytree.MTCredential {
	for _, credential := range gm.Credentials {
		if credential.ID == credentialID {
			return &credential
		}
	}

	return nil
}

func findOrCreateAccount(ps *pocketsmith.Client, userID int, instituationName string, accountName string, accountType moneytree.MTAccountType) (*pocketsmith.Account, error) {
	account, err := ps.FindAccountByName(userID, accountName)
	if err != nil {
		if err != pocketsmith.ErrNotFound {
			return nil, err
		}

		institution, err := ps.FindInstitutionByName(userID, instituationName)
		if err != nil {
			if err != pocketsmith.ErrNotFound {
				return nil, err
			}

			institution, err = ps.CreateInstitution(userID, instituationName, "jpy")
			if err != nil {
				return nil, err
			}
		}

		// check if there is an account in the institution
		// instAccounts, err := ps.GetInstitutionAccounts(institution.ID)
		// if err != nil {
		// 	return nil, err
		// }

		// if len(instAccounts) > 0 {
		// 	return &instAccounts[0], nil
		// }

		var psAccountType pocketsmith.AccountType
		switch accountType {
		case moneytree.MTAccountTypeBank:
			psAccountType = pocketsmith.AccountTypeBank
		case moneytree.MTAccountTypeCreditCard:
			psAccountType = pocketsmith.AccountTypeCredits
		case moneytree.MTAccountTypeStoredValue:
			psAccountType = pocketsmith.AccountTypeOtherAsset
		case moneytree.MTAccountTypePoint:
			psAccountType = pocketsmith.AccountTypeOtherAsset
		default:
			psAccountType = pocketsmith.AccountTypeOtherAsset
		}

		account, err := ps.CreateAccount(userID, institution.ID, accountName, "jpy", psAccountType)
		if err != nil {
			return nil, err
		}

		return account, nil
	}

	return account, nil
}

func main() {
	config := getConfig()

	ps := pocketsmith.NewClient(config.PocketsmithToken)
	currentUserRes, err := ps.GetCurrentUser()
	if err != nil {
		panic(err)
	}

	mt := moneytree.NewClient(config.MoneytreeApiKey)
	_, err = mt.GetAccessToken(config.MoneytreeUsername, config.MoneytreePassword)
	if err != nil {
		panic(err)
	}

	guestMeta, err := mt.GetGuestMeta()
	if err != nil {
		panic(err)
	}

	mt.RefreshAllCredentials()

	accounts, err := mt.GetAccounts()
	if err != nil {
		panic(err)
	}

	for _, account := range accounts {
		if account.Status == "closed" || account.Currency != "JPY" {
			continue
		}

		if account.AccountType != moneytree.MTAccountTypeBank && account.AccountType != moneytree.MTAccountTypeCreditCard && account.AccountType != moneytree.MTAccountTypeStoredValue {
			continue
		}

		fmt.Println("Processing moneytree account: ", account.InstitutionAccountName)

		credential := findCredentialFromMeta(guestMeta, account.CredentialID)
		if credential == nil {
			fmt.Println("Credential not found for account: ", account.InstitutionAccountName)
			continue
		}

		psAccount, err := findOrCreateAccount(ps, currentUserRes.ID, credential.InstitutionName, account.Nickname, account.AccountType)
		if err != nil {
			fmt.Println("Error creating account: ", err)
			panic(err)
		}

		// update account balance
		updateRes, err := ps.UpdateTransactionAccount(psAccount.PrimaryTransactionAccount.ID, psAccount.PrimaryTransactionAccount.Institution.ID, float64(account.CurrentBalance), time.Now().Format("2006-01-02"))
		if err != nil {
			fmt.Println("Error updating account balance: ", err)
			panic(err)
		}

		fmt.Println("updated account balance: ", updateRes.CurrentBalance)

		page := 1
		var mergedTxs []*moneytree.MTTransaction
		for {
			txs, err := mt.GetTransactions(account.ID, "2000-01-01", page, 500)
			if err != nil {
				fmt.Println("Error getting transactions: ", err)
				panic(err)
			}

			if len(txs) == 0 {
				break
			}

			mergedTxs = append(mergedTxs, txs...)

			page++
		}

		fmt.Println("num merged txs: ", len(mergedTxs))

		repeatedFoundTransactions := 0
		for _, tx := range mergedTxs {
			if repeatedFoundTransactions > 15 {
				fmt.Println("Too many repeated transactions found, likely everything processed already. Skipping...")
				break
			}

			name := tx.DescriptionPretty
			if tx.DescriptionGuest != "" {
				name = tx.DescriptionGuest
			}

			name = strings.TrimSpace(name)
			convertedPayee := sanitizier.Sanitize(name)

			fmt.Printf("[%d/%d] Processing moneytree transaction: %d %s\n", repeatedFoundTransactions+1, len(mergedTxs), tx.ID, convertedPayee)

			// Convert to pocketsmith transaction
			mtidMemo := fmt.Sprintf("mtid=%d", tx.RawTransactionID)
			psTx := &pocketsmith.CreateTransaction{
				Payee:       convertedPayee,
				Amount:      tx.Amount,
				Date:        tx.Date.Format("2006-01-02"),
				IsTransfer:  strings.Contains(name, "振込"),
				NeedsReview: false,
				// Note:         fmt.Sprintf("%s %d", strings.TrimSpace(tx.DescriptionPretty), tx.ID),
				Memo:         fmt.Sprintf("%s %s", name, mtidMemo),
				ChequeNumber: fmt.Sprintf("%d", tx.RawTransactionID),
			}

			searchResByChequeNumber, err := ps.SearchTransactionsByMemoContains(psAccount.PrimaryTransactionAccount.ID, tx.Date, mtidMemo)
			if err != nil {
				fmt.Println("Error searching transactions by cheque number: ", err)
				continue
			}

			if len(searchResByChequeNumber) > 0 {
				fmt.Println("Found transaction by cheque number: ", name)
				repeatedFoundTransactions++
				continue
			}

			// try to find the transaction first
			searchRes, err := ps.SearchTransactions(psAccount.PrimaryTransactionAccount.ID, tx.Date.Format("2006-01-02"), tx.Date.Format("2006-01-02"), fmt.Sprintf("%d", tx.ID))
			if err != nil {
				fmt.Println("Error searching transactions: ", err)
				continue
			}

			if len(searchRes) > 0 {
				updated := false
				for _, tx := range searchRes {
					// check if memo is set, if not, it's an older transaction and we need to upsert it
					if tx.Memo == "" {
						fmt.Println("memo not set, updating transaction to new format", name)
						if strings.Contains(psTx.Note, fmt.Sprintf("%d", tx.ID)) {
							psTx.Note = ""
						}

						err = ps.UpdateTransaction(tx.ID, psTx)
						if err != nil {
							fmt.Println("Error updating", err)
							continue
						}

						fmt.Println("Updated transaction: ", tx.ID)
						updated = true
					}
				}

				if updated {
					continue
				}

				fmt.Println("Found transaction already, won't add it again: ", name)
				repeatedFoundTransactions++
				continue
			}

			_, err = ps.AddTransaction(psAccount.PrimaryTransactionAccount.ID, psTx)
			if err != nil {
				fmt.Println("Error adding transaction: ", err)
				panic(err)
			}
		}
	}
}
