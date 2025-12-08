package server

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"udp/utils"

	"go.uber.org/zap"
)

var logger = utils.GetLogger()

func ToUppercase(data string) string {
	return strings.ToUpper(data)
}

/*
	Atividade 1: Equipe 10
	Implemente um sistema distribuído cliente-servidor usando Sockets TCP que gerencie
	um dicionário de termos (poucos exemplos, e.g., 5) e definições compartilhado em memória. O servi-
	dor deve aceitar múltiplas conexões e permitir que clientes consultem, insiram e atualizem definições.

	O TCP é usado para garantir a entrega ordenada dos comandos e a integridade das modificações.
	O cliente pode consultar a definição de um termo usando o comando LOOKUP <termo>. Para
	modificar o dicionário, o cliente pode inserir um novo termo (INSERT <termo> <definição>), ou
	modificar o dicionário, o cliente falhando se o termo já existir, ou modificar a definição de um termo
	existente usando o comando UPDATE <termo> <nova_definição>. É fundamental implementar
	mecanismos de bloqueio (locks ou mutexes) no servidor para garantir que a modificação de um termo
	por um cliente não seja interrompida ou sobreposta por outro cliente que tente modificar o mesmo
	termo simultaneamente, assegurando a integridade transacional dos dados.
*/

func ProcessDictCommand(request *utils.HTTPRequest, dict *Dictionary, mux *sync.Mutex) utils.HTTPResponse {
	startTime := time.Now()
	var response utils.HTTPResponse

	defer func() {
		elapsed := time.Since(startTime)
		logger.Info("Processed command",
			zap.String("method", request.Method),
			zap.String("path", request.Path),
			zap.Int("status_code", response.StatusCode),
			zap.Int64("elapsed_time", elapsed.Nanoseconds()))
		logger.Info("Response", zap.Int("status_code", response.StatusCode), zap.String("message", response.Message))
		time.Sleep(5 * time.Nanosecond)
	}()

	command := request.Method
	term := request.Path

	switch command {
	case "LIST":
		for !mux.TryLock() {
			if time.Since(startTime) > 30*time.Second {
				response = utils.HTTPResponse{
					StatusCode: http.StatusRequestTimeout,
					Message:    "Timeout while trying to access dictionary",
				}
				return response
			}
		}
		terms := dict.List()
		mux.Unlock()
		keys := "[" + strings.Join(terms, ", ") + "]"

		response = utils.HTTPResponse{
			StatusCode: http.StatusOK,
			Message:    keys,
		}
		return response

	case "LOOKUP":
		for !mux.TryLock() {
			if time.Since(startTime) > 30*time.Second {
				response = utils.HTTPResponse{
					StatusCode: http.StatusRequestTimeout,
					Message:    "Timeout while trying to access dictionary",
				}
				return response
			}
		}
		definition, exists := dict.LookUp(term)
		mux.Unlock()

		if !exists {
			response = utils.HTTPResponse{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("Term '%s' not found", term),
			}
			return response
		}

		response = utils.HTTPResponse{
			StatusCode: http.StatusOK,
			Message:    definition,
		}
		return response

	case "INSERT":
		if request.Body == "" {
			response = utils.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Message:    "INSERT command requires a body (definition)",
			}
			return response
		}

		for !mux.TryLock() {
			if time.Since(startTime) > 30*time.Second {
				response = utils.HTTPResponse{
					StatusCode: http.StatusRequestTimeout,
					Message:    "Timeout while trying to access dictionary",
				}
				return response
			}
		}
		defer mux.Unlock()

		success := dict.Insert(term, request.Body)

		if !success {
			response = utils.HTTPResponse{
				StatusCode: http.StatusConflict,
				Message:    fmt.Sprintf("Term '%s' already exists", term),
			}
			return response
		}

		response = utils.HTTPResponse{
			StatusCode: http.StatusCreated,
			Message:    fmt.Sprintf("Term '%s' inserted successfully", term),
		}
		return response

	case "UPDATE":
		if request.Body == "" {
			response = utils.HTTPResponse{
				StatusCode: http.StatusBadRequest,
				Message:    "UPDATE command requires a body (new definition)",
			}
			return response
		}

		for !mux.TryLock() {
			if time.Since(startTime) > 30*time.Second {
				response = utils.HTTPResponse{
					StatusCode: http.StatusRequestTimeout,
					Message:    "Timeout while trying to access dictionary",
				}
				return response
			}
		}
		defer mux.Unlock()

		success := dict.Update(term, request.Body)

		if !success {
			response = utils.HTTPResponse{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("Term '%s' does not exist", term),
			}
			return response
		}

		response = utils.HTTPResponse{
			StatusCode: http.StatusOK,
			Message:    fmt.Sprintf("Term '%s' updated successfully", term),
		}
		return response

	default:
		response = utils.HTTPResponse{
			StatusCode: http.StatusNotImplemented,
			Message:    fmt.Sprintf("Unknown command '%s'. Try one of: LIST, LOOKUP, INSERT, UPDATE", command),
		}
		return response
	}
}
