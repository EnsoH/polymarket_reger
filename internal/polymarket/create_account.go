package polymarket

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"polymarket/internal/web3"
	"polymarket/utils"
	"time"
)

func CreateAccount(pc *Client, wc *web3.Wallet, ampCookie string) error {
	currTime := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	expTime := time.Now().UTC().Add(7 * 24 * time.Hour).Format("2006-01-02T15:04:05.000Z")

	polyNonce, nonce, err := pc.GetNonce()
	if err != nil {
		return err
	}

	msgToSign := fmt.Sprintf(`polymarket.com wants you to sign in with your Ethereum account:
%s

Welcome to Polymarket! Sign to connect.

URI: https://polymarket.com
Version: 1
Chain ID: 137
Nonce: %s
Issued At: %s
Expiration Time: %s`, wc.Address, nonce, currTime, expTime)

	signature, err := wc.SignMsg(msgToSign)
	if err != nil {
		return err
	}

	token, err := createBearerToken(wc, nonce, currTime, expTime, signature)
	if err != nil {
		return err
	}

	proxyAddress, err := wc.CreateProxyAddress()
	if err != nil {
		return err
	}

	polySession, err := pc.Login(polyNonce, token, ampCookie)
	if err != nil {
		return err
	}

	data, err := pc.CreateProfile(proxyAddress, wc, ampCookie, polyNonce, polySession)
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	pc.PutPreferencec(data.Users[0].Preferences[0].ID, polyNonce, polySession)

	firstName := utils.GenerateRandomName()

	pc.PutFirstName(data.ID, firstName, polyNonce, polySession)

	time.Sleep(5 * time.Second)

	msgToSignTrade := MsgTradingEnable()
	tradeOnSig, _ := wc.SignTypedMsg(msgToSignTrade)

	pc.EnableTrading(wc, tradeOnSig, proxyAddress, polyNonce, polySession)

	return nil
}

func createBearerToken(wc *web3.Wallet, nonce, currTime, expTime, signature string) (string, error) {
	bearer := BearerToken{
		Address:        wc.Address.String(),
		ChainID:        137,
		Nonce:          nonce,
		Domain:         "polymarket.com",
		IssuedAt:       currTime,
		ExpirationTime: expTime,
		URI:            "https://polymarket.com",
		Statement:      "Welcome to Polymarket! Sign to connect.",
		Version:        "1",
	}

	dataB, err := json.Marshal(bearer)
	if err != nil {
		fmt.Println(err)
	}

	mergeString := fmt.Sprintf("%s:::%s", string(dataB), signature)
	data := base64.StdEncoding.EncodeToString([]byte(mergeString))

	return data, nil
}

func MsgTradingEnable() string {
	return `{
		"types": {
			"CreateProxy": [
				{
					"name": "paymentToken",
					"type": "address"
				},
				{
					"name": "payment",
					"type": "uint256"
				},
				{
					"name": "paymentReceiver",
					"type": "address"
				}
			],
			"EIP712Domain": [
				{
					"name": "name",
					"type": "string"
				},
				{
					"name": "chainId",
					"type": "uint256"
				},
				{
					"name": "verifyingContract",
					"type": "address"
				}
			]
		},
		"primaryType": "CreateProxy",
		"domain": {
			"name": "Polymarket Contract Proxy Factory",
			"chainId": "137",
			"verifyingContract": "0xaacfeea03eb1561c4e67d661e40682bd20e3541b"
		},
		"message": {
			"paymentToken": "0x0000000000000000000000000000000000000000",
			"payment": "0",
			"paymentReceiver": "0x0000000000000000000000000000000000000000"
		}
	}`
}
