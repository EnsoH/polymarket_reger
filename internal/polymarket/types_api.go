package polymarket

import "time"

type AmpPayload struct {
	DeviceID      string `json:"deviceId"`
	SessionID     int64  `json:"sessionId"`
	OptOut        bool   `json:"optOut"`
	LastEventTime int64  `json:"lastEventTime"`
	LastEventID   int    `json:"lastEventId"`
}

type NonceResp struct {
	Nonce string `json:"nonce"`
}

type BearerToken struct {
	Address        string `json:"address"`
	ChainID        int    `json:"chainId"`
	Nonce          string `json:"nonce"`
	Domain         string `json:"domain"`
	IssuedAt       string `json:"issuedAt"`
	ExpirationTime string `json:"expirationTime"`
	URI            string `json:"uri"`
	Statement      string `json:"statement"`
	Version        string `json:"version"`
}

type ProfilePayloadCreate struct {
	DisplayUsernamePublic bool   `json:"displayUsernamePublic"`
	EmailOptIn            bool   `json:"emailOptIn"`
	Name                  string `json:"name"`
	ProxyWallet           string `json:"proxyWallet"`
	Pseudonym             string `json:"pseudonym"`
	Referral              string `json:"referral"`
	UtmCampaign           string `json:"utmCampaign"`
	UtmContent            string `json:"utmContent"`
	UtmMedium             string `json:"utmMedium"`
	UtmSource             string `json:"utmSource"`
	UtmTerm               string `json:"utmTerm"`
	WalletActivated       bool   `json:"walletActivated"`
	Users                 []struct {
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
	} `json:"users"`
}

type RespCreateProfile struct {
	ID                    string    `json:"id"`
	Name                  string    `json:"name"`
	User                  int       `json:"user"`
	Referral              string    `json:"referral"`
	CreatedAt             time.Time `json:"createdAt"`
	UtmSource             string    `json:"utmSource"`
	UtmMedium             string    `json:"utmMedium"`
	UtmCampaign           string    `json:"utmCampaign"`
	UtmContent            string    `json:"utmContent"`
	UtmTerm               string    `json:"utmTerm"`
	WalletActivated       bool      `json:"walletActivated"`
	Pseudonym             string    `json:"pseudonym"`
	DisplayUsernamePublic bool      `json:"displayUsernamePublic"`
	Sync                  bool      `json:"_sync"`
	ProxyWallet           string    `json:"proxyWallet"`
	Users                 []struct {
		ID             string    `json:"id"`
		Username       string    `json:"username"`
		Provider       string    `json:"provider"`
		Blocked        bool      `json:"blocked"`
		CreatedAt      time.Time `json:"createdAt"`
		ProfileID      int       `json:"profileID"`
		Address        string    `json:"address"`
		ProxyWallet    string    `json:"proxyWallet"`
		IsExternalAuth bool      `json:"isExternalAuth"`
		Creator        bool      `json:"creator"`
		Mod            bool      `json:"mod"`
		Sync           bool      `json:"_sync"`
		Preferences    []struct {
			ID                           string `json:"id"`
			MarketInterests              string `json:"marketInterests"`
			EmailNotificationPreferences string `json:"emailNotificationPreferences"`
			AppNotificationPreferences   string `json:"appNotificationPreferences"`
			UserID                       int    `json:"userID"`
			PreferencesStatus            string `json:"preferencesStatus"`
			SubscriptionStatus           bool   `json:"subscriptionStatus"`
			Sync                         bool   `json:"_sync"`
		} `json:"preferences"`
		WalletPreferences []struct {
			ID                      string `json:"id"`
			GasPreference           string `json:"gasPreference"`
			WalletPreferencesStatus string `json:"walletPreferencesStatus"`
			CustomGasPrice          string `json:"customGasPrice"`
			UserID                  int    `json:"userID"`
			AdvancedMode            bool   `json:"advancedMode"`
			Sync                    bool   `json:"_sync"`
		} `json:"walletPreferences"`
	} `json:"users"`
	IsCloseOnly bool `json:"isCloseOnly"`
}

type Preferences struct {
	EmailNotificationPreferences string `json:"emailNotificationPreferences"`
	MarketInterests              string `json:"marketInterests"`
}

type PutFirstName struct {
	DisplayUsernamePublic bool   `json:"displayUsernamePublic"`
	Name                  string `json:"name"`
	Referral              string `json:"referral"`
}

type PayloadEnableTrading struct {
	From            string `json:"from"`
	To              string `json:"to"`
	ProxyWallet     string `json:"proxyWallet"`
	Data            string `json:"data"`
	Signature       string `json:"signature"`
	SignatureParams struct {
		PaymentToken    string `json:"paymentToken"`
		Payment         string `json:"payment"`
		PaymentReceiver string `json:"paymentReceiver"`
	} `json:"signatureParams"`
	Type string `json:"type"`
}

type RespEnableTrade struct {
	TransactionID   string `json:"transactionID"`
	TransactionHash string `json:"transactionHash"`
	State           string `json:"state"`
}

type RespTxCheck []struct {
	TransactionID   string    `json:"transactionID"`
	TransactionHash string    `json:"transactionHash"`
	From            string    `json:"from"`
	To              string    `json:"to"`
	ProxyAddress    string    `json:"proxyAddress"`
	Data            string    `json:"data"`
	Nonce           string    `json:"nonce"`
	Value           string    `json:"value"`
	Signature       string    `json:"signature"`
	State           string    `json:"state"`
	Type            string    `json:"type"`
	Owner           string    `json:"owner"`
	Metadata        string    `json:"metadata"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
