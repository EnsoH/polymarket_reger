package web3

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

type Wallet struct {
	Client     *ethclient.Client
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

func New(rpc string, privateKey string) (*Wallet, error) {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil, fmt.Errorf("can't connect to RPC: %w", err)
	}

	pKey, err := crypto.HexToECDSA(privateKey[2:])
	if err != nil {
		return nil, fmt.Errorf("can't convert private key: %w", err)
	}

	wallet := &Wallet{
		Client:     client,
		PrivateKey: pKey,
		Address:    crypto.PubkeyToAddress(pKey.PublicKey),
	}

	return wallet, nil
}

func (w *Wallet) GetNonce(ctx context.Context) (uint64, error) {
	nonce, err := w.Client.PendingNonceAt(ctx, w.Address)
	if err != nil {
		return 0, fmt.Errorf("error getting nonce: %w", err)
	}

	return nonce, err
}

func (w *Wallet) GetTipCap(ctx context.Context) (*big.Int, error) {
	tipCap, err := w.Client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting gas tip: %w", err)
	}

	return tipCap, nil
}

func (w *Wallet) GetBaseFee(ctx context.Context) (*big.Int, error) {
	header, err := w.Client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting block header: %w", err)
	}

	return header.BaseFee, nil
}

func (w *Wallet) GetMaxFeePerGas(ctx context.Context) (*big.Int, error) {
	tipCap, err := w.GetTipCap(ctx)
	if err != nil {
		return nil, err
	}

	baseFee, err := w.GetBaseFee(ctx)
	if err != nil {
		return nil, err
	}

	maxFeePerGas := new(big.Int).Add(baseFee, tipCap)

	return maxFeePerGas, nil
}

func (w *Wallet) GetGasLimit(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	gasLimit, err := w.Client.EstimateGas(ctx, msg)
	if err != nil {
		return 0, fmt.Errorf("error when estimate gas limit: %w", err)
	}

	return gasLimit, nil
}

func (w *Wallet) SignMsg(data string) (string, error) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	hash := crypto.Keccak256Hash([]byte(msg))

	signature, err := crypto.Sign(hash.Bytes(), w.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("error when sign signature: %w", err)
	}

	hexSignature := hexutil.Encode(signature)

	return hexSignature, nil
}

func (w *Wallet) CreateProxyAddress() (string, error) {
	addressType, err := abi.NewType("address", "", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create address type: %w", err)
	}

	arguments := abi.Arguments{
		{
			Type: addressType,
		},
	}

	abiEncoded, err := arguments.Pack(w.Address)
	if err != nil {
		return "", fmt.Errorf("failed to ABI encode address: %w", err)
	}

	hash := crypto.Keccak256(abiEncoded)

	safeFactoryAddress := common.HexToAddress("0xaacFeEa03eb1561C4e67d661e40682Bd20E3541b")
	safeInitCodeHash := common.FromHex("0x2bce2127ff07fb632d16c8347c4ebf501f4841168bed00d9e6ef715ddb6fcecf")

	create2Address := crypto.CreateAddress2(
		safeFactoryAddress,
		common.BytesToHash(hash),
		safeInitCodeHash,
	)

	return create2Address.String(), nil
}

func (w *Wallet) SignTypedMsg(data string) (string, error) {
	var tData apitypes.TypedData
	err := json.Unmarshal([]byte(data), &tData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal typed data: %v", err)
	}

	typedDataHash, err := tData.HashStruct(tData.PrimaryType, tData.Message)
	if err != nil {
		return "", fmt.Errorf("failed to hash typed data: %v", err)
	}

	domainSeparator, err := tData.HashStruct("EIP712Domain", tData.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("failed to hash typed data: %v", err)
	}

	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	challenge := crypto.Keccak256(rawData)
	s, err := crypto.Sign(challenge, w.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %v", err)
	}
	s[64] += 27

	hexSignature := hexutil.Encode(s)

	return hexSignature, nil
}
