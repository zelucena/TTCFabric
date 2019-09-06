package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

type Timestamp time.Time

type Candidato struct {
	nome string
	email string
}

type Votacao struct {
	ID string `json:"id"`
	inicioCandidatura *Timestamp
	terminoCandidatura *Timestamp
	inicioVotacao *Timestamp
	terminoVotacao *Timestamp
}

type Votante struct {

}

type Voto struct {
	votante *Votante
	horario *Timestamp
	candidato *Candidato
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger
	if function == "cadastrarVotacao" {
		return s.cadastrarVotacao(APIstub, args)
	} else if function == "visualizarVotacao" {
		return s.visualizarVotacao(APIstub, args)
	} else if function == "cadastrarCandidato" {
		return s.cadastrarCandidato(APIstub, args)
	} else if function == "visualizarCandidatos" {
		return s.visualizarCandidatos(APIstub)
	} else if function == "votar" {
		return s.votar(APIstub, args)
	}

	return shim.Error("Função indisponível.")
}

//estilo 1, recebendo objeto
func (s *SmartContract) cadastrarVotacao(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 5{
		return shim.Error("Esperados 5 parâmetros: ID, início candidatura, término candidatura, início votação, término votação")
	}
	objetoJSON := args[0]
	var votacao = Votacao{
		args[0], args[1], args[2], args[3], args[4]
	}

	//verifica unicidade
	val, erro := stub.GetState(votacao.ID)
	if val != nil {
		return shim.Error(fmt.Sprintf("%s", erro))
	}

	if erro = stub.PutState(votacao.ID, []byte(objetoJSON)); erro != nil {
		mensagemErro := fmt.Sprintf("Erro: não é possível inserir votação com id <%d>, devido a %s", votacao.ID, erro)
		fmt.Println(mensagemErro)
		return shim.Error(mensagemErro)
	}

	return shim.Success(nil)
}

//estilo 2, recebendo lista
func (s *SmartContract) visualizarVotacao(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	votacaoAsBytes, _ := APIstub.GetState(args[0])
	if votacaoAsBytes == nil {
		return shim.Error("Não foi possível localizar votação")
	}
	return shim.Success(votacaoAsBytes)
}

