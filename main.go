package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/core/chaincode/shim/ext/cid"
	"github.com/hyperledger/fabric/protos/peer"
	"regexp"
	"sort"
	"time"
)

//formatos de data
const (
	ISO_DATE = "2006-01-02 15:04:05"
	BR_DATE  = "02/01/2006 15:04:05"
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

//metodo para ordenaçao dos candidatos vencedores
// ByNumeroVotos implementa sort.Interface baseado no campo Candidato.Votos
type ByNumeroVotos []Candidato
func (a ByNumeroVotos) Len() int           { return len(a) }
func (a ByNumeroVotos) Less(i, j int) bool { return a[i].NumeroVotos > a[j].NumeroVotos }
func (a ByNumeroVotos) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (s *VotacaoContract) Init(APIstub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

//retorna o estado atual da votacao
func (s *VotacaoContract) getVotacao(APIstub shim.ChaincodeStubInterface) (Votacao, error) {
	var votacao = Votacao{}
	var state, GetStateError = APIstub.GetState("votacao")

	if GetStateError != nil {
		return Votacao{}, GetStateError
	}

	if state != nil {
		var unmarshalErro = json.Unmarshal(state, &votacao)
		if unmarshalErro != nil {
			return Votacao{}, unmarshalErro
		}
	}
	return votacao, nil
}

//funcao local para validacao de formato de emails, conforme necessidade de validar parametros de entrada
func (s *VotacaoContract) validarEmail(email string) (bool, error){
	Re, erroCompile := regexp.Compile(`^[a-z0-9._\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

	if erroCompile != nil {
		return false, erroCompile
	}
	return Re.MatchString(email), nil
}

//afere atraves do certificado se o client e do tipo Admin@org.suffix
func (s *VotacaoContract) isAdmin(APIstub shim.ChaincodeStubInterface) (bool, error){
	var certificado, erroCertificado = cid.GetX509Certificate(APIstub)
	if erroCertificado != nil {
		return false, erroCertificado
	}

	Re, erroCompile := regexp.Compile(`^Admin@[a-z]+.[a-z]+$`)

	if erroCompile != nil {
		return false, erroCompile
	}

	return Re.MatchString(certificado.Subject.CommonName), nil
}

//afere atraves do certificado se o client e do tipo User{n}@org.suffix
func (s *VotacaoContract) isUser(APIstub shim.ChaincodeStubInterface) (bool, error){
	var certificado, erroCertificado = cid.GetX509Certificate(APIstub)
	if erroCertificado != nil {
		return false, erroCertificado
	}

	Re, erroCompile := regexp.Compile(`^User[0-9]+@[a-z]+.[a-z]+$`)

	if erroCompile != nil {
		return false, erroCompile
	}

	return Re.MatchString(certificado.Subject.CommonName), nil
}

//gera um ID unico para o usuario
func (s *VotacaoContract) getClientHash(APIstub shim.ChaincodeStubInterface) (string, error) {
	clientID, erroID := cid.GetID(APIstub)
	clientMSPID, erroMSPID := cid.GetMSPID(APIstub)

	if erroID != nil {
		return "", erroID
	}

	if erroMSPID != nil {
		return "", erroMSPID
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(clientMSPID + clientID))), nil
}

//Recebe o nome da funcao e os parametros de entrada
//trata atributos do client
//chama a funcao desejada, validando permissoes
func (s *VotacaoContract) Invoke(APIstub shim.ChaincodeStubInterface) peer.Response {
	// Extrair funçao e parâmetros chamados
	function, args := APIstub.GetFunctionAndParameters()

	clientHash, erroHash := s.getClientHash(APIstub)
	if erroHash != nil {
		return shim.Error(erroHash.Error())
	}

	isUser, erroUser := s.isUser(APIstub)
	if erroUser != nil {
		return shim.Error(erroUser.Error())
	}

	isAdmin, erroAdmin := s.isAdmin(APIstub)
	if erroAdmin != nil {
		return shim.Error(erroAdmin.Error())
	}

	// invoca a funçao apropriada
	if function == "cadastrarVotacao" {
		if !isAdmin {
			return shim.Error("Funcao exclusiva para o administrador")
		}
		return s.cadastrarVotacao(APIstub, args)
	} else if function == "visualizarVotacao" {
		return s.visualizarVotacao(APIstub, args)
	} else if function == "cadastrarCandidato" {
		if !isAdmin {
			return shim.Error("Funcao exclusiva para o administrador")
		}
		return s.cadastrarCandidato(APIstub, args)
	} else if function == "visualizarCandidatos" {
		return s.visualizarCandidatos(APIstub, args)
	} else if function == "votar" {
		if !isUser {
			return shim.Error("Funcao exclusiva para usuarios")
		}
		return s.votar(APIstub, args, clientHash)
	} else if function == "visualizarVoto" {
		if !isUser {
			return shim.Error("Funcao exclusiva para usuarios")
		}
		return s.visualizarVoto(APIstub, args, clientHash)
	} else if function == "divulgarResultados" {
		return s.divulgarResultados(APIstub, args)
	}

	return shim.Error("Funcao indisponivel.")
}

/**
assumindo a existência de apenas uma votaçao por canal, portanto dentro de uma chaincode, apenas um objeto de votaçao
O objeto de votaçao pode ser editado contando que nao haja votos ou candidatos
Nao permite colisao entre periodo de cadastro de candidatos e votacao
 */
func (s *VotacaoContract) cadastrarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	//validar parâmetros de entrada
	if len(args) != 4 {
		return shim.Error("Parâmetros esperados: inicio candidatura, termino candidatura, inicio votacao, termino votacao")
	}

	var inicioCandidatura, 	erro1 = time.Parse(BR_DATE, args[0])
	var terminoCandidatura, erro2 = time.Parse(BR_DATE, args[1])
	var inicioVotacao, 		erro3 = time.Parse(BR_DATE, args[2])
	var terminoVotacao, 	erro4 = time.Parse(BR_DATE, args[3])

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
		return shim.Error("O inicio das candidaturas deve ser uma data anterior ao termino das candidaturas")
	}

	if inicioVotacao.Equal(terminoVotacao) || inicioVotacao.After(terminoVotacao) {
		return shim.Error("O inicio das candidaturas deve ser uma data anterior ao termino das candidaturas")
	}

	if terminoCandidatura.Equal(inicioVotacao) || terminoCandidatura.After(inicioVotacao) {
		return shim.Error("As candidaturas devem encerrar antes do inicio da votacao")
	}

	var votacao, erroVotacao	= s.getVotacao(APIstub)

	if erroVotacao != nil {
		return shim.Error(erroVotacao.Error())
	}

	if len(votacao.Candidatos) > 0 {
		return shim.Error("Nao e possivel alterar a votaçao, ja existem candidatos computados")
	}

	if len(votacao.Votos) > 0 {
		return shim.Error("Nao e possivel alterar a votacao, ja existem votos computados")
	}

	votacao = Votacao{
		ObjectType:         "votacao",
		ID:                 "votacao",
		InicioCandidatura:  inicioCandidatura.Format(ISO_DATE),
		TerminoCandidatura: terminoCandidatura.Format(ISO_DATE),
		InicioVotacao:      inicioVotacao.Format(ISO_DATE),
		TerminoVotacao:     terminoVotacao.Format(ISO_DATE),
		Candidatos:         make(map[string]Candidato),
		Votos:              make(map[string]Voto),
	}

	var votacaoAsBytes, erroJSON = json.Marshal(votacao)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}

	var putStateError = APIstub.PutState("votacao", votacaoAsBytes)

	if putStateError != nil {
		return shim.Error(putStateError.Error())
	}

	return shim.Success(nil)
}

//cadastra candidato com nome e email, validando limite de 50 caracteres para nome (arbitrario) e 254 para email (RFC 3696)
//disponivel no periodo valido de cadastro de candidatos
func (s *VotacaoContract) cadastrarCandidato(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var dataAtual = time.Now()
	if len(args) != 3 {
		return shim.Error("Esperado: ID, nome, email")
	}

	var votacao, erroVotacao = s.getVotacao(APIstub)

	if erroVotacao != nil {
		return shim.Error(erroVotacao.Error())
	}

	if votacao.InicioCandidatura == "" || votacao.TerminoCandidatura == "" {
		return shim.Error("Nao ha uma votacao em curso")
	}

	var inicioCandidatura,  erroFormatoInicio = time.Parse(ISO_DATE, votacao.InicioCandidatura)

	if erroFormatoInicio != nil {
		return shim.Error(erroFormatoInicio.Error())
	}

	var terminoCandidatura, erroFormatoFim = time.Parse(ISO_DATE, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	if inicioCandidatura.After(dataAtual) {
		return shim.Error("O periodo de candidaturas ira comecar em " + inicioCandidatura.Format(BR_DATE))
	}

	if terminoCandidatura.Before(dataAtual) {
		return shim.Error("O periodo de candidaturas ja terminou em " + terminoCandidatura.Format(BR_DATE))
	}

	email := args[2]
	if (len(email) > 254) {
		return shim.Error("Email do candidato nao pode exceder 254 caracteres")
	}

	validacaoEmail, erroValidacaoEmail := s.validarEmail(email)
	if erroValidacaoEmail != nil {
		return shim.Error(erroValidacaoEmail.Error())
	}
	if !validacaoEmail {
		return shim.Error("Formato invalido de email")
	}

	nomeCandidato := args[1]

	if (len(nomeCandidato) > 50) {
		return shim.Error("Nome do candidato nao pode exceder 50 caracteres")
	}

	var candidato = Candidato{
		ObjectType: "candidato",
		ID:          args[0],
		Nome:        nomeCandidato,
		Email:       email,
		NumeroVotos: 0,
	}

	for _, v := range votacao.Candidatos {
		if v.ID == candidato.ID {
			return shim.Error("ID ja cadastrado")
		}

		if v.Email == candidato.Email {
			return shim.Error("Email ja cadastrado")
		}
	}

	votacao.Candidatos[candidato.ID] = candidato

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

// audita todas as modificacoes na votacao, expondo todas as informacoes publicamente. Disponivel apos o termino da votacao
func (s *VotacaoContract) visualizarVotacao(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votacao, erro = s.getVotacao(APIstub)
	if erro != nil {
		return shim.Error(erro.Error())
	}

	if votacao.InicioVotacao == "" || votacao.TerminoVotacao == "" {
		return shim.Error("Nao ha uma votacao em curso")
	}

	var terminoVotacao, erroFormatoFim = time.Parse(ISO_DATE, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	var dataAtual = time.Now()
	if terminoVotacao.After(dataAtual) {
		return shim.Error("O periodo de votacao encerra em "+terminoVotacao.Format(BR_DATE))
	}

	historyIterator, erroGetHistory := APIstub.GetHistoryForKey("votacao")

	if erroGetHistory != nil {
		return shim.Error(erroGetHistory.Error())
	}

	var historico []string

	for historyIterator.HasNext() {
		modificacao, erroHistory := historyIterator.Next()
		if erroHistory != nil {
			return shim.Error(erroHistory.Error())
		}
		historico = append(historico, string(modificacao.Value))
	}

	var historicoAsBytes, erroJSON = json.Marshal(historico)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(historicoAsBytes)
}

//retorna apenas a lista de votos. Disponivel apos a votacao ter encerrado
func (s *VotacaoContract) visualizarVotos(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votacao, erro = s.getVotacao(APIstub)

	if erro != nil {
		return shim.Error(erro.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(ISO_DATE, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	var dataAtual = time.Now()
	if terminoVotacao.After(dataAtual) {
		return shim.Error("O periodo de votacao se encerra em " + terminoVotacao.Format(BR_DATE))
	}

	var votosAsBytes, erroJSON = json.Marshal(votacao.Votos)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(votosAsBytes)
}

//realiza a contagem de votos na rede apos o periodo de encerramento
func (s *VotacaoContract) divulgarResultados(APIstub shim.ChaincodeStubInterface, args []string) peer.Response {
	var votacao, erro = s.getVotacao(APIstub)

	if erro != nil {
		return shim.Error(erro.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(ISO_DATE, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	var dataAtual = time.Now()
	if terminoVotacao.After(dataAtual) {
		return shim.Error("O periodo de votacao se encerra em " + terminoVotacao.Format(BR_DATE))
	}

	//garantir que os votos estao zerados
	var candidatos = votacao.Candidatos
	for id, candidato := range candidatos {
		temp := candidato
		temp.NumeroVotos = 0
		candidatos[id] = temp
	}

	//contar votos
	for _, voto := range votacao.Votos {
		candidato := candidatos[voto.Candidato.ID]
		candidato.NumeroVotos++
		candidatos[voto.Candidato.ID] = candidato
	}

	var candidatosSlice []Candidato

	for _, candidato := range candidatos {
		candidatosSlice = append(candidatosSlice, candidato)
	}
	sort.Sort(ByNumeroVotos(candidatosSlice))
	var votosAsBytes, erroJSON = json.Marshal(candidatosSlice)

	if erroJSON != nil {
		return shim.Error(erroJSON.Error())
	}
	return shim.Success(votosAsBytes)
}

//visualiza o proprio voto, independente do termino da votacao
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

//visualiza todos os candidatos cadastrados
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

//realiza um voto unico dentro de um periodo valido de votacoes. Informar ID do candidato
func (s *VotacaoContract) votar(APIstub shim.ChaincodeStubInterface, args []string, clientHash string) peer.Response {
	dataAtual		:= time.Now()
	candidatoID		:= args[0]

	var votacao, erroVotacao = s.getVotacao(APIstub)

	if erroVotacao != nil {
		return shim.Error(erroVotacao.Error())
	}

	if votacao.InicioVotacao == "" || votacao.TerminoVotacao == "" {
		return shim.Error("Nao ha uma votacao em curso")
	}

	var inicioVotacao,  erroFormatoInicio = time.Parse(ISO_DATE, votacao.InicioVotacao)

	if erroFormatoInicio != nil {
		return shim.Error(erroFormatoInicio.Error())
	}

	var terminoVotacao, erroFormatoFim = time.Parse(ISO_DATE, votacao.TerminoVotacao)

	if erroFormatoFim != nil {
		return shim.Error(erroFormatoFim.Error())
	}

	if inicioVotacao.After(dataAtual) {
		return shim.Error("O periodo de votacao comeca em "+inicioVotacao.Format(BR_DATE))
	}

	if terminoVotacao.Before(dataAtual) {
		return shim.Error("O periodo de votacao encerrou em "+terminoVotacao.Format(BR_DATE))
	}

	if len(args) != 1 {
		return shim.Error("Esperado: ID do candidato")
	}

	var voto = Voto{}
	if votacao.Votos[clientHash] != voto {
		return shim.Error("Nao e permitido votar duas vezes")
	}

	var candidatoBranco = Candidato{}
	if votacao.Candidatos[candidatoID] == candidatoBranco {
		return shim.Error("Candidato invalido")
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