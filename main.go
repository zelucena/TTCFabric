package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/tree/release-1.4/core/chaincode/lib/cid"
	"sort"
	"time"
)


type Candidato struct {
	ObjectType	string  `json:"doctype"`
	ID			string  `json:"id"`
	Nome		string  `json:"nome"`
	Email		string  `json:"email"`
	NumeroVotos int		`json:"numerovotos"`
}

type Votacao struct {
	ObjectType		   string 				`json:"doctype"`
	ID                 string 				`json:"id"`
	InicioCandidatura  string 				`json:"iniciocandidatura"`
	TerminoCandidatura string 				`json:"terminocandidatura"`
	InicioVotacao      string 				`json:"iniciovotacao"`
	TerminoVotacao     string 				`json:"terminovotacao"`
	Candidatos		map[string]Candidato 	`json:"candidatos"`
	Votos			map[string]Voto			`json:"votos"`
}

type Votante struct {
	ObjectType 	string 		`json:"doctype"`
	ID			string		`json:"id"`
}

type Voto struct {
	ObjectType 	string 		`json:"doctype"`
	Votante		Votante 	`json:"votante"`
	Timestamp 	string 		`json:"timestamp"`
	Candidato 	Candidato 	`json:"candidato"`
}

//Classe da chaincode
type VotacaoContract struct { }

// ByNumeroVotos implementa sort.Interface baseado no campo Candidato.Votos
type ByNumeroVotos []Candidato
func (a ByNumeroVotos) Len() int           { return len(a) }
func (a ByNumeroVotos) Less(i, j int) bool { return a[i].NumeroVotos > a[j].NumeroVotos }
func (a ByNumeroVotos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (s *VotacaoContract) InitVotacao(APIstub shim.ChaincodeStubInterface) peer.Response {
	var votacao, GetStateError = APIstub.getState("votacao")

	if GetStateError != nil {
		return shim.Error(GetStateError.Error())
	}

	if votacao == nil {
		votacao = Votacao{
			ObjectType:        	"votacao",
			ID:                	"votacao",
			Candidatos:		   	make(map[string]Candidato),
			Votos:				make(map[string]Voto),
		}
	}

	var retornoJSON, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}

	var putStateError = APIstub.PutState("votacao", retornoJSON)

	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	return shim.Success(nil)
}

func (s *VotacaoContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	return s.InitVotacao(APIstub)
}

func (s *VotacaoContract) getVotacao(APIstub shim.ChaincodeStubInterface) (Votacao, error) {
	var state, GetStateError = APIstub.getState("votacao")

	if GetStateError != nil {
		return Votacao{}, GetStateError
	}

	var votacao = Votacao{}
	var unmarshalErro = json.Unmarshal(state, &votacao)
	if unmarshalErro != nil {
		return Votacao{}, unmarshalErro
	}
	return votacao, nil
}

func (s *VotacaoContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {
	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()

	clientID, erroID := cid.GetID(APIstub)
	clientMSPID, erroMSPID := cid.GetMSPID(APIstub)

	if erroID != nil {
		return shim.Error(erroID.Error())
	}

	if erroMSPID != nil {
		return shim.Error(erroID.Error())
	}
	clientHash	:= fmt.Sprintf("%x", sha256.Sum256([]byte(clientMSPID + clientID)))

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
		return s.votar(APIstub, args, clientHash)
	} else if function == "visualizarVoto" {
		return s.visualizarVoto(APIstub, args, clientHash)
	} else if function == "divulgarResultados" {
		return s.divulgarResultados(APIstub, args)
	}

	return shim.Error("Funcao indisponivel.")
}

/**
Vamos assumir a existência de apenas uma votação por canal, portanto dentro de uma chaincode, apenas um objeto de votação
O objeto de votação pode ser editado contanto que não haja votos
 */
func (s *VotacaoContract) cadastrarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	//definir formato de entrada
	formatoData := "2006-01-02 15:04:05"

	//validar parâmetros de entrada
	if len(args) != 4 {
		return shim.Error("Parâmetros esperados: inicio candidatura, termino candidatura, inicio votacao, termino votacao")
	}

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

	if inicioCandidatura.Equal(terminoCandidatura) || inicioCandidatura.After(terminoCandidatura) {
		return shim.Error("O início das candidaturas deve ser uma data anterior ao término das candidaturas")
	}

	if inicioVotacao.Equal(terminoVotacao) || inicioVotacao.After(terminoVotacao) {
		return shim.Error("O início das candidaturas deve ser uma data anterior ao término das candidaturas")
	}

	state, getStateError := APIstub.GetState("votacao")

	if getStateError != nil {
		return shim.Error(getStateError.Error())
	}

	var votacao = Votacao{}

	if state != nil {
		var unmarshalErro = json.Unmarshal(state, &votacao)
		if unmarshalErro != nil {
			return shim.Error(unmarshalErro.Error())
		}

		if leng(votacao.Votos) > 0 {
			return shim.Error("Não é possível alterar a votação, já existem votos computados")
		}
	} else {
		votacao.ObjectType			= "votacao"
		votacao.ID 					= "votacao"
	}

	votacao.InicioCandidatura 	= inicioCandidatura.Format(formatoData)
	votacao.TerminoCandidatura 	= terminoCandidatura.Format(formatoData)
	votacao.InicioVotacao 		= inicioVotacao.Format(formatoData)
	votacao.TerminoVotacao 		= terminoVotacao.Format(formatoData)

	var votacaoAsBytes, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}

	var putStateError = APIstub.PutState(votacao.ID, votacaoAsBytes)

	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	return shim.Success(nil)
}

func (s *VotacaoContract) cadastrarCandidato(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	formatoData := "2006-01-02 15:04:05"
	var dataAtual = time.Now()
	if len(args) != 3 {
		return shim.Error("Esperado: ID, nome, email")
	}

	var votacao, erroVotacao = s.getVotacao(APIstub)

	if erroVotacao != nil {
		return shim.Error(erroVotacao.Error())
	}

	var inicioCandidatura,  erroFormatoInicio = time.Parse(formatoData, votacao.InicioCandidatura)

	if erroFormatoInicio != nil {
		return shim.Error(erroFormatoInicio.Error())
	}

	var terminoCandidatura, erroFormatoFim = time.Parse(formatoData, votacao.InicioCandidatura)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	if inicioCandidatura.After(dataAtual) {
		return shim.Error("O periodo de candidaturas ainda não comecou")
	}

	if terminoCandidatura.Before(dataAtual) {
		return shim.Error("O periodo de candidaturas ja terminou")
	}

	var candidato = Candidato{
		ObjectType: "candidato",
		ID:          args[0],
		Nome:        args[1],
		Email:       args[2],
		NumeroVotos: 0,
	}

	for _, v := range votacao.Candidatos {
		if v.ID == candidato.ID {
			return shim.Error("ID já inserido")
		}

		if v.Email == candidato.Email {
			return shim.Error("Email já inserido")
		}
	}

	votacao.Candidatos["id"] = candidato

	var votacaoAsBytes, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}

	var putStateError = APIstub.PutState(votacao.ID, votacaoAsBytes)

	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	return shim.Success(nil)
}

func (s *VotacaoContract) visualizarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	formatoData := "2006-01-02 15:04:05"
	var votacao, erro = s.getVotacao(APIstub)
	if erro != nil {
		return shim.Error(erro.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(formatoData, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	var dataAtual = time.Now()
	if terminoVotacao.After(dataAtual) {
		return shim.Error("O periodo de votacao ainda não encerrou")
	}

	var votacaoAsBytes, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(votacaoAsBytes)
}

func (s *VotacaoContract) visualizarVotos(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	formatoData := "2006-01-02 15:04:05"
	var votacao, erro = s.getVotacao(APIstub)

	if erro != nil {
		return shim.Error(erro.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(formatoData, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	var dataAtual = time.Now()
	if terminoVotacao.After(dataAtual) {
		return shim.Error("O período de votação ainda não encerrou")
	}

	var votosAsBytes, erroJSON = json.Marshal(votacao.Votos)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(votosAsBytes)
}

func (s *VotacaoContract) divulgarResultados(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	formatoData := "2006-01-02 15:04:05"
	var votacao, erro = s.getVotacao(APIstub)

	if erro != nil {
		return shim.Error(erro.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(formatoData, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	var dataAtual = time.Now()
	if terminoVotacao.After(dataAtual) {
		return shim.Error("O periodo de votacao ainda não encerrou")
	}

	candidatos := make(map[string]*Candidato)
	for id, candidato := range votacao.Candidatos {
		candidatos[id] = &candidato
		candidatos[id].NumeroVotos = 0
	}

	for _, voto := range votacao.Votos {
		candidatos[voto.Candidato.ID].NumeroVotos++
	}

	var candidatosSlice []Candidato

	for _, candidato := range candidatos {
		candidatosSlice = candidatosSlice.append(candidato)
	}
	sort.Sort(ByNumeroVotos(candidatosSlice))
	var votosAsBytes, erroJSON = json.Marshal(candidatosSlice)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(votosAsBytes)
}

func (s *VotacaoContract) visualizarVoto(APIstub shim.ChaincodeStubInterface, args []string, clientHash string) peer.Response {
	var votacao, erro = s.getVotacao(APIstub)

	if erro != nil {
		return shim.Error(erro.Error())
	}

	var voto = Voto{}
	if votacao.Votos[clientHash] == voto {
		return shim.Error("Este cliente ainda nao votou")
	}

	var votoAsBytes, erroJSON = json.Marshal(votacao.Votos[clientHash])

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(votoAsBytes)
}

func (s *VotacaoContract) visualizarCandidatos(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votacao, erro = s.getVotacao(APIstub)

	if erro != nil {
		return shim.Error(erro.Error())
	}

	var candidatosAsBytes, erroJSON = json.Marshal(votacao.Candidatos)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(candidatosAsBytes)
}

func (s *VotacaoContract) votar(APIstub shim.ChaincodeStubInterface, args []string, clientHash string) peer.Response {
	formatoData 	:= "2006-01-02 15:04:05"
	dataAtual		:= time.Now()
	candidatoID		:= args[0]

	var votacao, erroVotacao = s.getVotacao(APIstub)

	if erroVotacao != nil {
		return shim.Error(erroVotacao.Error())
	}

	var inicioVotacao,  erroFormatoInicio = time.Parse(formatoData, votacao.InicioVotacao)

	if erroFormatoInicio != nil {
		return shim.Error(erroFormatoInicio.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(formatoData, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	if inicioVotacao.After(dataAtual) {
		return shim.Error("O periodo de candidaturas ainda não comecou")
	}

	if terminoVotacao.Before(dataAtual) {
		return shim.Error("O periodo de candidaturas ja terminou")
	}

	if len(args) != 1 {
		return shim.Error("Esperado: ID do candidato")
	}

	var voto = Voto{}
	if votacao.Votos[clientHash] != voto {
		return shim.Error("Não é permitido votar duas vezes")
	}

	var candidatoBranco = Candidato{}
	if votacao.Candidatos[candidatoID] == candidatoBranco {
		return shim.Error("Candidato inválido")
	}

	var horarioTransacao, erroTimestamp = APIstub.GetTxTimestamp()
	if erroTimestamp != nil {
		return shim.Error(erroTimestamp.Error())
	}

	voto.ObjectType	= "voto"
	voto.Votante 	= Votante{
		ObjectType: "votante",
		ID:         clientHash,
	}

	voto.Timestamp  = horarioTransacao.String()
	voto.Candidato  = votacao.Candidatos[candidatoID]

	votacao.Votos[clientHash] = voto

	var votacaoAsBytes, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}

	var putStateError = APIstub.PutState(votacao.ID, votacaoAsBytes)

	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(VotacaoContract))
	if err != nil {
		fmt.Printf(err.Error())
	}
}