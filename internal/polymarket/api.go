package polymarket

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"polymarket/internal/web3"
	"polymarket/utils/client"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/google/uuid"
)

type Client struct {
	Client *http.Client
}

func New() *Client {
	return &Client{
		Client: client.New(),
	}
}

func (c *Client) ParseCookie(prefix string, cookies []*http.Cookie) (string, error) {
	var polymarketNonce string
	found := false

	for _, cookie := range cookies {
		if strings.HasPrefix(prefix, cookie.Name) {
			polymarketNonce = cookie.Value
			found = true
			break
		}
	}

	if !found {
		return "", errors.New("no cookie found with prefix")
	}

	return polymarketNonce, nil
}

func (c *Client) GenerateAMPCookie() string {
	u := uuid.New()

	payload := AmpPayload{
		DeviceID:      u.String(),
		SessionID:     time.Now().UnixMilli(),
		OptOut:        false,
		LastEventTime: time.Now().Add(15 * time.Second).UnixMilli(),
		LastEventID:   0,
	}

	bodyBytes, _ := json.Marshal(payload)
	data := base64.StdEncoding.EncodeToString(bodyBytes)

	return data
}

func (c *Client) GetNonce() (string, string, error) {
	log.Info("Get Nonce")
	const ApiEndpoint = "https://gamma-api.polymarket.com/nonce"

	req, _ := http.NewRequest(http.MethodGet, ApiEndpoint, nil)

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("cookie", "AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE")
	req.Header.Set("origin", "https://polymarket.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error when do request to get nonce: %w", err)
	}

	defer resp.Body.Close()

	var nonce NonceResp

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error read body nonce: %w", err)
	}

	err = json.Unmarshal(body, &nonce)
	if err != nil {
		return "", "", fmt.Errorf("error when unmarshal body nonce: %w", err)
	}

	polymarketNonce, err := c.ParseCookie("polymarketnonce=", resp.Cookies())
	if err != nil {
		return "", "", fmt.Errorf("error parsing cookies: %w", err)
	}

	return polymarketNonce, nonce.Nonce, nil
}

func (c *Client) Login(polyNonce, token, ampCookie string) (string, error) {
	const ApiEndpoint = "https://gamma-api.polymarket.com/login"

	req, _ := http.NewRequest("GET", ApiEndpoint, nil)

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("cookie", fmt.Sprintf("AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE; polymarketnonce=%s; AMP_4572e28e5c=%s", polyNonce, ampCookie))
	req.Header.Set("origin", "https://polymarket.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error when request login: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		polymarketSession, err := c.ParseCookie("polymarketsession=", resp.Cookies())
		if err != nil {
			return "", fmt.Errorf("error not found cookie polymarketsession: %w", err)
		}

		return polymarketSession, nil
	}

	return "", err
}

func (c *Client) CreateProfile(proxyAddress string, walletAddress *web3.Wallet, ampCookie, polyNonce, polySession string) (*RespCreateProfile, error) {
	log.Printf("Creating profile %v", walletAddress.Address.String())

	payload := &ProfilePayloadCreate{
		DisplayUsernamePublic: true,
		EmailOptIn:            false,
		Name:                  fmt.Sprintf("%s-%d", proxyAddress, time.Now().UnixMilli()),
		ProxyWallet:           proxyAddress,
		Pseudonym:             proxyAddress,
		Referral:              "",
		UtmCampaign:           "",
		UtmContent:            "",
		UtmMedium:             "",
		UtmSource:             "",
		UtmTerm:               "",
		WalletActivated:       false,
		Users: []struct {
			Address        string `json:"address"`
			IsExternalAuth bool   `json:"isExternalAuth"`
			Provider       string `json:"provider"`
			ProxyWallet    string `json:"proxyWallet"`
			Username       string `json:"username"`
			Preferences    []struct {
				EmailNotificationPreferences string `json:"emailNotificationPreferences"`
				AppNotificationPreferences   string `json:"appNotificationPreferences"`
				MarketInterests              string `json:"marketInterests"`
				PreferencesStatus            string `json:"preferencesStatus"`
				SubscriptionStatus           bool   `json:"subscriptionStatus"`
			} `json:"preferences"`
			WalletPreferences []struct {
				AdvancedMode            bool   `json:"advancedMode"`
				CustomGasPrice          string `json:"customGasPrice"`
				GasPreference           string `json:"gasPreference"`
				WalletPreferencesStatus string `json:"walletPreferencesStatus"`
			} `json:"walletPreferences"`
		}{
			{
				Address:        walletAddress.Address.String(),
				IsExternalAuth: true,
				Provider:       "metamask",
				ProxyWallet:    proxyAddress,
				Username:       fmt.Sprintf("%s-%d", proxyAddress, time.Now().UnixMilli()),
				Preferences: []struct {
					EmailNotificationPreferences string `json:"emailNotificationPreferences"`
					AppNotificationPreferences   string `json:"appNotificationPreferences"`
					MarketInterests              string `json:"marketInterests"`
					PreferencesStatus            string `json:"preferencesStatus"`
					SubscriptionStatus           bool   `json:"subscriptionStatus"`
				}{
					{
						EmailNotificationPreferences: `{"generalEmail":{"sendEmails":false},"marketEmails":{"sendEmails":false},"newsletterEmails":{"sendEmails":false},"promotionalEmails":{"sendEmails":false},"eventEmails":{"sendEmails":false,"tagIds":["2","21","1","107","596","74"]},"orderFillEmails":{"sendEmails":false,"hideSmallFills":true},"resolutionEmails":{"sendEmails":false}}`,
						AppNotificationPreferences:   `{"eventApp":{"sendApp":true,"tagIds":["2","21","1","107","596","74"]},"marketPriceChangeApp":{"sendApp":true},"orderFillApp":{"sendApp":true,"hideSmallFills":true},"resolutionApp":{"sendApp":true}}`,
						MarketInterests:              "[]",
						PreferencesStatus:            "New/Existing - Created Prefs",
						SubscriptionStatus:           false,
					},
				},
				WalletPreferences: []struct {
					AdvancedMode            bool   `json:"advancedMode"`
					CustomGasPrice          string `json:"customGasPrice"`
					GasPreference           string `json:"gasPreference"`
					WalletPreferencesStatus string `json:"walletPreferencesStatus"`
				}{
					{
						AdvancedMode:            false,
						CustomGasPrice:          "30",
						GasPreference:           "fast",
						WalletPreferencesStatus: "New/Existing - Created Wallet Prefs",
					},
				},
			},
		},
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "https://gamma-api.polymarket.com/profiles", bytes.NewBuffer(body))

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", fmt.Sprintf("AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE; polymarketnonce=%s; AMP_4572e28e5c=%s; polymarketsession=%s; polymarketauthtype=metamask", polyNonce, ampCookie, polySession))
	req.Header.Set("origin", "https://polymarket.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error when request create profile: %w", err)
	}

	defer resp.Body.Close()

	var data *RespCreateProfile

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error when read body create profile: %w", err)
	}

	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshal create profile body: %w", err)
	}

	log.Printf("Profile created | %v", walletAddress.Address.String())

	return data, nil
}

func (c *Client) PutPreferencec(preferenceId, polyNonce, polySession string) {
	log.Print("Put Preference")

	apiEndpoint := fmt.Sprintf("https://gamma-api.polymarket.com/preferences/%s", preferenceId)

	payload := Preferences{
		EmailNotificationPreferences: `{"generalEmail":{"sendEmails":true},"marketEmails":{"sendEmails":true},"newsletterEmails":{"sendEmails":true},"promotionalEmails":{"sendEmails":true},"eventEmails":{"sendEmails":true,"tagIds":["2","21","1","107","596","74"]},"orderFillEmails":{"sendEmails":true,"hideSmallFills":true},"resolutionEmails":{"sendEmails":true}}`,
		MarketInterests:              "[]",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, apiEndpoint, bytes.NewBuffer(body))
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", fmt.Sprintf("AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE; polymarketnonce=%s; polymarketsession=%s; polymarketauthtype=metamask; AMP_4572e28e5c=JTdCJTdE", polyNonce, polySession))
	req.Header.Set("origin", "https://polymarket.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Print("GOOD")
	}

}

func (c *Client) PutFirstName(userId, firstName, polyNonce, polySession string) {
	log.Print("Put First Name")

	apiEndpoint := fmt.Sprintf("https://gamma-api.polymarket.com/profiles/%s", userId)

	payload := PutFirstName{
		DisplayUsernamePublic: true,
		Name:                  firstName,
		Referral:              "",
	}

	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, apiEndpoint, bytes.NewBuffer(body))

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", fmt.Sprintf("AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE; polymarketnonce=%s; polymarketsession=%s; polymarketauthtype=metamask; AMP_4572e28e5c=JTdCJTdE", polyNonce, polySession))
	req.Header.Set("origin", "https://polymarket.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		log.Print("GOOD")
	}
}

func (c *Client) EnableTrading(wallet *web3.Wallet, signature, proxyAddress, polyNonce, polySession string) {
	log.Print("enabling trading...")
	const apiEndpoint = "https://relayer-v2.polymarket.com/submit"

	payload := PayloadEnableTrading{
		From:        wallet.Address.String(),
		To:          "0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b",
		ProxyWallet: proxyAddress,
		Data:        "0x",
		Signature:   signature,
		SignatureParams: struct {
			PaymentToken    string `json:"paymentToken"`
			Payment         string `json:"payment"`
			PaymentReceiver string `json:"paymentReceiver"`
		}{
			PaymentToken:    "0x0000000000000000000000000000000000000000",
			Payment:         "0",
			PaymentReceiver: "0x0000000000000000000000000000000000000000",
		},
		Type: "SAFE-CREATE",
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(bodyBytes)
	}

	req, _ := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewBuffer(bodyBytes))

	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("cookie", fmt.Sprintf("AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE; polymarketnonce=%s; polymarketsession=%s; polymarketauthtype=metamask; AMP_4572e28e5c=JTdCJTdE", polyNonce, polySession))
	req.Header.Set("origin", "https://polymarket.com")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-site")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

	resp, err := c.Client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var data RespEnableTrade
	_ = json.Unmarshal(body, &data)

	if resp.StatusCode == 200 {
		fmt.Println(data.TransactionID, data.State)
		_ = c.CheckTxStatus(data.TransactionID, polyNonce, polySession)
		log.Print("success")
	}
}

func (c *Client) CheckTxStatus(txID, polyNonce, polySession string) error {
	for i := 0; i < 5; i++ {
		const apiEndpoint = "https://relayer-v2.polymarket.com/transaction?id="

		fmt.Println(apiEndpoint + txID)

		req, _ := http.NewRequest(http.MethodGet, apiEndpoint+txID, nil)

		req.Header.Set("accept", "application/json, text/plain, */*")
		req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Set("content-type", "application/json")
		req.Header.Set("cookie", fmt.Sprintf("AMP_MKTG_4572e28e5c=JTdCJTIycmVmZXJyZXIlMjIlM0ElMjJodHRwcyUzQSUyRiUyRnd3dy5nb29nbGUuY29tJTJGJTIyJTJDJTIycmVmZXJyaW5nX2RvbWFpbiUyMiUzQSUyMnd3dy5nb29nbGUuY29tJTIyJTdE; polymarketnonce=%s; polymarketsession=%s; polymarketauthtype=metamask; AMP_4572e28e5c=JTdCJTdE", polyNonce, polySession))
		req.Header.Set("origin", "https://polymarket.com")
		req.Header.Set("priority", "u=1, i")
		req.Header.Set("sec-ch-ua", `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`)
		req.Header.Set("sec-ch-ua-mobile", "?0")
		req.Header.Set("sec-ch-ua-platform", `"Windows"`)
		req.Header.Set("sec-fetch-dest", "empty")
		req.Header.Set("sec-fetch-mode", "cors")
		req.Header.Set("sec-fetch-site", "same-site")
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")

		time.Sleep(2 * time.Second)

		resp, err := c.Client.Do(req)
		if err != nil {
			_ = fmt.Errorf("error when request to check tx %w", err)
		}

		fmt.Println(resp.StatusCode)

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			_ = fmt.Errorf("error when read body tx %w", err)
		}

		var data RespTxCheck

		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println(err)
		}

		if resp.StatusCode == 200 {
			if data[0].State == "STATE_MINED" {
				log.Printf("Tx is mined %s", data[0].State)
				break
			} else {
				log.Printf("Tx is not mined %s", data[0].State)
				time.Sleep(5 * time.Second)
			}
		}
	}

	return errors.New("tx check failed")
}
