package main

import (
	"bytes"
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
	ID                 string `json:"id"`
	InicioCandidatura  string `json:"iniciocandidatura"`
	TerminoCandidatura string `json:"terminocandidatura"`
	InicioVotacao      string `json:"iniciovotacao"`
	TerminoVotacao     string `json:"terminovotacao"`
}

type Votante struct {

}

type Voto struct {
	Assinatura	string `json:"votante"`
	Timestamp 	string `json:"timestamp"`
	Candidato 	Candidato `json:"candidato"`
}

type queryResponse struct {
	Key        string
	Value      string
	Namespace  string
}

func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string)([] byte, error) {
	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)
	resultsIterator, err := stub.GetQueryResult(queryString)
	defer resultsIterator.Close()
	if err != nil {
		return nil, err
	}
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse,
			err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
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
	} else if function == "addTeste" {
		return s.addTeste(APIstub, args)
	} else if function == "queryTeste" {
		return s.queryTeste(APIstub, args)
	} else if function == "getSignedProposal" {
		return s.getSignedProposal(APIstub, args)
	} else if function == "getCreator" {
		return s.getCreator(APIstub, args)
	} else if function == "auditarVotos" {
		return s.auditarVotos(APIstub, args)
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

	var votacao = Votacao{}
	votacao.ID = ID
	votacao.InicioCandidatura = inicioCandidatura.Format(formatoData)
	votacao.TerminoCandidatura = terminoCandidatura.Format(formatoData)
	votacao.InicioVotacao = inicioVotacao.Format(formatoData)
	votacao.TerminoVotacao = terminoVotacao.Format(formatoData)

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
	var Voto = Voto{}

	//var creator, erroCreator = APIstub.GetCreator()
	//if erroCreator != nil {
	//	return shim.Error(fmt.Sprintf("%s", erroCreator))
	//}

	var horarioTransacao, erroTimestamp = APIstub.GetTxTimestamp()
	if erroTimestamp != nil {
		return shim.Error(fmt.Sprintf("%s", erroTimestamp))
	}

	//Voto.Assinatura = fmt.Sprintf("%s", creator)
	Voto.Assinatura = horarioTransacao.String()
	Voto.Timestamp  = horarioTransacao.String()
	Voto.Candidato  = Candidato{}
	Voto.Candidato.email 	= "email_teste@ttcfabric.com"
	Voto.Candidato.nome		= "John Doe"

	var VotoAsBytes, erroJSON = json.Marshal(Voto)

	if erroJSON != nil {
		return shim.Error(fmt.Sprintf("%s", erroJSON))
	}

	var putStateError = APIstub.PutState(Voto.Assinatura, VotoAsBytes)

	if putStateError != nil {
		return shim.Error(fmt.Sprintf("%s", putStateError))
	}

	return shim.Success(nil)
}

func (s *VotacaoContract) addTeste(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votacao = Votacao{}

	votacao.ID = "teste"
	votacao.InicioCandidatura = "2019-01-01 10:00:00"
	votacao.TerminoCandidatura = "2019-01-08 23:00:00"
	votacao.InicioVotacao = "2019-07-01 10:00:00"
	votacao.TerminoVotacao = "2019-07-01 23:00:00"

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

func (s *VotacaoContract) queryTeste(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votacao = Votacao{}

	votacao.ID = "teste"
	votacao.InicioCandidatura = "2019-01-01 10:00:00"
	votacao.TerminoCandidatura = "2019-01-08 23:00:00"
	votacao.InicioVotacao = "2019-07-01 10:00:00"
	votacao.TerminoVotacao = "2019-07-01 23:00:00"
	//
	var votacaoAsBytes, _ = json.Marshal(votacao)

	return shim.Success(votacaoAsBytes)
}

func (s *VotacaoContract) getSignedProposal(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var proposal, erroAPI = APIstub.GetSignedProposal()

	if erroAPI != nil {
		return shim.Error(fmt.Sprintf("%s", erroAPI))
	}

	var retornoJSON, erroJSON = json.Marshal(proposal)

	if erroJSON != nil {
		return shim.Error(fmt.Sprintf("%s", erroJSON))
	}

	return shim.Success(retornoJSON)
}

func (s *VotacaoContract) getCreator(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var creator, erro = APIstub.GetCreator()
	if erro != nil {
		return shim.Error(fmt.Sprintf("%s", erro))
	}
	return shim.Success(creator)
}

func (s *VotacaoContract) auditarVotos(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votos, erroConsulta = getQueryResultForQueryString(APIstub, "")
	if erroConsulta != nil {
		return shim.Error(fmt.Sprintf("%s", erroConsulta))
	}
	return shim.Success(votos)
}

func main() {
	err := shim.Start(new(VotacaoContract))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}