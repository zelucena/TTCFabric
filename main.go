package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"time"
)


type Candidato struct {
	nome string
	email string
}

type Votacao struct {
	ID string
	inicioCandidatura string
	terminoCandidatura string
	inicioVotacao string
	terminoVotacao string
	cadastro string
}

type Votante struct {

}

type Voto struct {
	votante Votante
	horario string
	candidato Candidato
}

//Esta é a classe da chaincode
type VotacaoContract struct { }

func (s *VotacaoContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (s *VotacaoContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

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
		return s.visualizarCandidatos(APIstub, args)
	} else if function == "votar" {
		return s.votar(APIstub, args)
	}

	return shim.Error("Funcao indisponivel.")
}

//estilo 1, recebendo objeto
func (s *VotacaoContract) cadastrarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	formatoData := "2006-01-02 15:04:05"
	if len(args) != 5 {
		return shim.Error("Esperados 6 parametros: Metodo, ID, inicio candidatura, termino candidatura, inicio votacao, termino votacao")
	}

	var ID 						  = args[0]
	var inicioCandidatura, 	erro1 = time.Parse(formatoData, args[1])
	var terminoCandidatura, erro2 = time.Parse(formatoData, args[2])
	var inicioVotacao, 		erro3 = time.Parse(formatoData, args[3])
	var terminoVotacao, 	erro4 = time.Parse(formatoData, args[4])

	if erro1 != nil {
		return shim.Error(erro1.Error())
	}

	if erro2 != nil {
		return shim.Error(erro2.Error())
	}

	if erro3 != nil {
		return shim.Error(erro3.Error())
	}

	if erro4 != nil {
		return shim.Error(erro4.Error())
	}

	horarioTransacao,_ := stub.GetTxTimestamp()
	horarioTransacao = time.Unix(timestamp.Seconds, int64(timestamp.Nanos)).String()
	var votacao = Votacao{ID, inicioCandidatura.Format(formatoData), terminoCandidatura.Format(formatoData), inicioVotacao.Format(formatoData), terminoVotacao.Format(formatoData), horarioTransacao}

	//verifica unicidade
	val, getStateError := APIstub.GetState(votacao.ID)
	if val != nil {
		return shim.Error(fmt.Sprintf("%s", "Erro: ID já existe"))
	}
	if getStateError != nil {
		return shim.Error(fmt.Sprintf("%s", getStateError))
	}

	var votacaoAsBytes, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(fmt.Sprintf("%s", erroJSON))
	}

	var putStateError = APIstub.PutState(votacao.ID, votacaoAsBytes)

	if putStateError != nil {
		mensagemErro := fmt.Sprintf("Erro: nao e possivel inserir votacao com id <%d>, devido a %s", votacao.ID, putStateError)
		fmt.Println(mensagemErro)
		return shim.Error(mensagemErro)
	}

	return shim.Success(nil)
}

func (s *VotacaoContract) cadastrarCandidato(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	return shim.Success(nil)
}

//estilo 2, recebendo lista
func (s *VotacaoContract) visualizarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	votacaoAsBytes, _ := APIstub.GetState(args[0])
	if votacaoAsBytes == nil {
		return shim.Error("Nao foi possivel localizar votacao")
	}
	return shim.Success(votacaoAsBytes)
}

func (s *VotacaoContract) visualizarCandidatos(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	return shim.Success(nil)
}

func (s *VotacaoContract) votar(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(VotacaoContract))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}