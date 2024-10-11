package main

import (
	"fmt"
	"polymarket/internal/polymarket"
	"polymarket/internal/web3"
)

func main() {
	const RPC = "https://rpc.ankr.com/polygon"
	const privateKey = "0x85e8f5ba193801d718f7c277f15f9a4c732bd061d96d8720fe606d2107ef8778"

	polyC := polymarket.New()
	wallet, err := web3.New(RPC, privateKey)
	if err != nil {
		fmt.Println(err)
	}

	ampCook := polyC.GenerateAMPCookie()

	err = polymarket.CreateAccount(polyC, wallet, ampCook)
	if err != nil {
		fmt.Println(err)
	}
}
