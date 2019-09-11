package chaincode

import (
	"bytes"
	"encoding/binary"
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
	inicioCandidatura time.Time
	terminoCandidatura time.Time
	inicioVotacao time.Time
	terminoVotacao time.Time
}

type Votante struct {

}

type Voto struct {
	votante *Votante
	horario *time.Time
	candidato *Candidato
}

type SmartContract struct {}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {

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

	return shim.Error("Função indisponível.")
}

//estilo 1, recebendo objeto
func (s *SmartContract) cadastrarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var buffer bytes.Buffer

	if len(args) != 5 {
		return shim.Error("Esperados 5 parâmetros: ID, início candidatura, término candidatura, início votação, término votação")
	}

	var ID 						  = args[0]
	var inicioCandidatura, 	erro1 = time.Parse(time.RFC3339, args[1])
	var terminoCandidatura, erro2 = time.Parse(time.RFC3339, args[2])
	var inicioVotacao, 		erro3 = time.Parse(time.RFC3339, args[3])
	var terminoVotacao, 	erro4 = time.Parse(time.RFC3339, args[4])

	if erro1 != nil {
		return shim.Error(erro1.Error())
	}

	if erro2 != nil {
		return shim.Error(erro1.Error())
	}

	if erro3 != nil {
		return shim.Error(erro1.Error())
	}

	if erro4 != nil {
		return shim.Error(erro1.Error())
	}

	var votacao = Votacao{
		ID, inicioCandidatura, terminoCandidatura, inicioVotacao, terminoVotacao,
	}

	//verifica unicidade
	val, getStateError := APIstub.GetState(votacao.ID)
	if val != nil {

	}
	if getStateError != nil {
		return shim.Error(fmt.Sprintf("%s", getStateError))
	}

	var bufferError = binary.Write(&buffer, binary.BigEndian, votacao)
	if bufferError != nil {
		return shim.Error(fmt.Sprintf("%s", bufferError))
	}

	var putStateError = APIstub.PutState(votacao.ID, buffer.Bytes())

	if putStateError != nil {
		mensagemErro := fmt.Sprintf("Erro: não é possível inserir votação com id <%d>, devido a %s", votacao.ID, putStateError)
		fmt.Println(mensagemErro)
		return shim.Error(mensagemErro)
	}

	return shim.Success(nil)
}

func (s *SmartContract) cadastrarCandidato(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	return shim.Success(nil)
}

//estilo 2, recebendo lista
func (s *SmartContract) visualizarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	votacaoAsBytes, _ := APIstub.GetState(args[0])
	if votacaoAsBytes == nil {
		return shim.Error("Não foi possível localizar votação")
	}
	return shim.Success(votacaoAsBytes)
}

func (s *SmartContract) visualizarCandidatos(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {

	return shim.Success(nil)
}

func (s *SmartContract) votar(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}