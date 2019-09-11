package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	chaincode "github.com/zelucena/TTCFabric"
)

func main() {
	err := shim.Start(new(chaincode.SmartContract))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}