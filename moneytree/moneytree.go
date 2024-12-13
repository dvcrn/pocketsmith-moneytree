package moneytree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GetAccessTokenResponse struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	ExpiresIn      int    `json:"expires_in"`
	RefreshToken   string `json:"refresh_token"`
	Scope          string `json:"scope"`
	CreatedAt      int    `json:"created_at"`
	ResourceServer string `json:"resource_server"`
}

type BalanceComponents struct {
	Available *float64 `json:"available"`
	Unclosed  *float64 `json:"unclosed"`
	Closed    *float64 `json:"closed"`
	Revolving *float64 `json:"revolving"`
}

type MTAccountType string

const (
	MTAccountTypeBank        MTAccountType = "bank"
	MTAccountTypeCreditCard  MTAccountType = "credit_card"
	MTAccountTypeStoredValue MTAccountType = "stored_value"
	MTAccountTypePoint       MTAccountType = "point"
)

type MTAccount struct {
	ID                       int               `json:"id"`
	GuestID                  int               `json:"guest_id"`
	Nickname                 string            `json:"nickname"`
	Currency                 string            `json:"currency"`
	CredentialID             int               `json:"credential_id"`
	AccountType              MTAccountType     `json:"account_type"`
	InstitutionAccountNumber string            `json:"institution_account_number"`
	InstitutionAccountName   string            `json:"institution_account_name"`
	BranchName               *string           `json:"branch_name"`
	Status                   string            `json:"status"`
	LastSuccessAt            string            `json:"last_success_at"`
	Group                    string            `json:"group"`
	DetailType               string            `json:"detail_type"`
	SubType                  string            `json:"sub_type"`
	CurrentBalance           float64           `json:"current_balance"`
	CurrentBalanceInBase     float64           `json:"current_balance_in_base"`
	BalanceComponents        BalanceComponents `json:"balance_components"`
}

type GetAccountsResponse struct {
	Accounts []MTAccount `json:"accounts"`
}

type MTTransaction struct {
	ID                     int           `json:"id"`
	Amount                 float64       `json:"amount"`
	Date                   time.Time     `json:"date"`
	DescriptionGuest       string        `json:"description_guest"`
	DescriptionPretty      string        `json:"description_pretty"`
	DescriptionRaw         string        `json:"description_raw"`
	RawTransactionID       int           `json:"raw_transaction_id"`
	AccountID              int           `json:"account_id"`
	ClaimID                int           `json:"claim_id"`
	CategoryID             int           `json:"category_id"`
	ExpenseType            int           `json:"expense_type"`
	PredictedExpenseType   int           `json:"predicted_expense_type"`
	CreatedAt              string        `json:"created_at"`
	UpdatedAt              string        `json:"updated_at"`
	TransactionAttachments []interface{} `json:"transaction_attachments"`
}

type GetTransactionsResponse struct {
	Transactions []*MTTransaction `json:"transactions"`
}

type MTPosition struct {
	ID                int      `json:"id"`
	Date              string   `json:"date"`
	NameRaw           string   `json:"name_raw"`
	NameClean         string   `json:"name_clean"`
	Quantity          float64  `json:"quantity"`
	MarketValue       float64  `json:"market_value"`
	AcquisitionValue  float64  `json:"acquisition_value"`
	Ticker            string   `json:"ticker"`
	Currency          string   `json:"currency"`
	AcctCurrencyValue *float64 `json:"acct_currency_value"`
	Profit            float64  `json:"profit"`
	AccountID         int      `json:"account_id"`
	Value             float64  `json:"value"`
	CostBasis         float64  `json:"cost_basis"`
}

type Moneytree struct {
	apiKey string

	accessToken  string
	refreshToken string
}

func NewClient(token string) *Moneytree {
	return &Moneytree{
		apiKey: token,
	}
}

func (m *Moneytree) GetAccessToken(guestLogin, password string) (*GetAccessTokenResponse, error) {
	url := "https://myaccount.getmoneytree.com/oauth/token"
	headers := map[string]string{
		"Accept":          "application/json",
		"Host":            "myaccount.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"User-Agent":      "Moneytree/1.16.3 (Android 12; en_AU; Pixel 3)",
		"Accept-Language": "en-US-POSIX",
		"locale":          "en-US-POSIX",
		"Content-Type":    "application/json; charset=utf-8",
	}

	body := map[string]string{
		"client_id":   m.apiKey,
		"grant_type":  "password",
		"guest_login": guestLogin,
		"password":    password,
	}

	resp, err := makeRequest("POST", url, headers, body)
	if err != nil {
		return nil, err
	}

	var result GetAccessTokenResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	m.accessToken = result.AccessToken
	m.refreshToken = result.RefreshToken

	return &result, nil
}

func (m *Moneytree) GetAccounts() ([]MTAccount, error) {
	url := "https://jp-api.getmoneytree.com/v8/api/accounts.json"
	headers := map[string]string{
		"Accept-Language": "en_AU",
		"locale":          "en_AU",
		"X-Api-Key":       m.apiKey,
		"X-Api-Version":   "20180814",
		"Accept":          "application/json",
		"User-Agent":      "Moneytree/1.16.3 (Android 12; en_AU; Pixel 3)",
		"Authorization":   "Bearer " + m.accessToken,
		"Host":            "jp-api.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/json",
	}

	resp, err := makeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	var result GetAccountsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Accounts, nil
}

func (m *Moneytree) GetTransactions(accountID int, since string, page, perPage int) ([]*MTTransaction, error) {
	url := fmt.Sprintf("https://jp-api.getmoneytree.com/v8/api/accounts/%d/transactions.json?since=%s&page=%d&per_page=%d",
		accountID, since, page, perPage)
	headers := map[string]string{
		"Accept-Language": "en_AU",
		"locale":          "en_AU",
		"X-Api-Key":       m.apiKey,
		"X-Api-Version":   "20180814",
		"Accept":          "application/json",
		"User-Agent":      "Moneytree/1.16.3 (Android 12; en_AU; Pixel 3)",
		"Authorization":   "Bearer " + m.accessToken,
		"Host":            "jp-api.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/json",
	}

	resp, err := makeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	var result GetTransactionsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Transactions, nil
}

func (m *Moneytree) GetPositions(accountID string) ([]MTPosition, error) {
	url := fmt.Sprintf("https://jp-api.getmoneytree.com/v8/api/accounts/%s/positions.json", accountID)
	headers := map[string]string{
		"Accept-Language": "en_AU",
		"locale":          "en_AU",
		"X-Api-Key":       m.apiKey,
		"X-Api-Version":   "20180814",
		"Accept":          "application/json",
		"User-Agent":      "Moneytree/1.16.3 (Android 12; en_AU; Pixel 3)",
		"Authorization":   "Bearer " + m.accessToken,
		"Host":            "jp-api.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/json",
	}

	resp, err := makeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	var result []MTPosition
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (m *Moneytree) RefreshAllCredentials() (interface{}, error) {
	url := "https://jp-api.getmoneytree.com/v8/api/credentials/refresh.json"
	headers := map[string]string{
		"Accept-Language": "en_AU",
		"locale":          "en_AU",
		"X-Api-Key":       m.apiKey,
		"X-Api-Version":   "20180814",
		"Accept":          "application/json",
		"User-Agent":      "Moneytree/1.16.3 (Android 12; en_AU; Pixel 3)",
		"Authorization":   "Bearer " + m.accessToken,
		"Host":            "jp-api.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/json",
	}

	resp, err := makeRequest("PUT", url, headers, nil)
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// makeRequest is a helper function that would need to be implemented
// to handle the actual HTTP requests
func makeRequest(method, url string, headers map[string]string, body interface{}) ([]byte, error) {

	client := &http.Client{}

	var req *http.Request
	var err error

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

type MTCategory struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	LocaleName  string `json:"locale_name"`
	ParentID    *int   `json:"parent_id"`
	ExpenseType int    `json:"expense_type"`
}

type GetCategoriesResponse struct {
	Categories []MTCategory `json:"categories"`
}

func (m *Moneytree) GetCategories() ([]MTCategory, error) {
	url := "https://jp-api.getmoneytree.com/v8/api/presenter/categories.json?locale=en"
	headers := map[string]string{
		"Accept-Language": "en-US,en;q=0.9",
		"X-Api-Key":       m.apiKey,
		"X-Api-Version":   "6",
		"Accept":          "application/json",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.2 Safari/605.1.15",
		"Authorization":   "Bearer " + m.accessToken,
		"Host":            "jp-api.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/json",
	}

	resp, err := makeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	var result GetCategoriesResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result.Categories, nil
}

type MTInstitution struct {
	ID int `json:"id"`
}

type MTCredential struct {
	ID                          int           `json:"id"`
	AdditionalStatusInformation interface{}   `json:"additional_status_information"`
	ErrorInfo                   interface{}   `json:"error_info"`
	LastSuccess                 string        `json:"last_success"`
	StatusSetAt                 string        `json:"status_set_at"`
	InstitutionName             string        `json:"institution_name"`
	BackgroundRefreshFrequency  int           `json:"background_refresh_frequency"`
	AuthType                    int           `json:"auth_type"`
	Status                      string        `json:"status"`
	AutoRun                     bool          `json:"auto_run"`
	UsesCertificate             bool          `json:"uses_certificate"`
	Institution                 MTInstitution `json:"institution"`
	Accounts                    []MTAccount   `json:"accounts"`
}

type MTGuest struct {
	ID                 int            `json:"id"`
	LocaleIdentifier   string         `json:"locale_identifier"`
	Email              string         `json:"email"`
	UnconfirmedEmail   interface{}    `json:"unconfirmed_email"`
	UID                string         `json:"uid"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	ConfirmationSentAt string         `json:"confirmation_sent_at"`
	ConfirmedAt        string         `json:"confirmed_at"`
	PaymentProvider    interface{}    `json:"payment_provider"`
	Country            string         `json:"country"`
	BaseCurrency       string         `json:"base_currency"`
	IntercomUserHash   string         `json:"intercom_user_hash"`
	SubscriptionLevel  string         `json:"subscription_level"`
	Credentials        []MTCredential `json:"credentials"`
}

type GetGuestResponse struct {
	Guest MTGuest `json:"guest"`
}

func (m *Moneytree) GetGuestMeta() (*MTGuest, error) {
	url := "https://jp-api.getmoneytree.com/v8/api/presenter/guests.json"
	headers := map[string]string{
		"Accept-Language": "en_AU",
		"locale":          "en_AU",
		"X-Api-Key":       m.apiKey,
		"X-Api-Version":   "20180814",
		"Accept":          "application/json",
		"User-Agent":      "Moneytree/1.16.3 (Android 12; en_AU; Pixel 3)",
		"Authorization":   "Bearer " + m.accessToken,
		"Host":            "jp-api.getmoneytree.com",
		"Connection":      "Keep-Alive",
		"Content-Type":    "application/json",
	}

	resp, err := makeRequest("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	var result GetGuestResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result.Guest, nil
}
