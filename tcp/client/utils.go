package client

import (
	"fmt"
	"strings"
	"tcp/utils"
)

func ToLowercase(data string) string {
	return strings.ToLower(data)
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

func ParseCommandToHTTPRequest(command string) (*utils.HTTPRequest, error) {
	parts := strings.Fields(command)
	if len(parts) < 1 {
		return nil, fmt.Errorf("command must have at least METHOD")
	}

	method := strings.ToUpper(parts[0])
	term := ""
	body := ""

	if method == "LIST" {
		term = ""
		body = ""
	} else {
		if len(parts) > 1 {
			term = parts[1]
		}
		if len(parts) > 2 {
			body = strings.Join(parts[2:], " ")
		}
	}

	request := &utils.HTTPRequest{
		Method: method,
		Path:   term,
		Body:   body,
	}

	return request, nil
}

func ParseHTTPResponse(response string) (statusCode int, statusText string, message string) {
	// utils.Logger.Info("Parsing HTTP response", zap.String("response", response))
	response = strings.TrimSpace(response)

	parts := strings.SplitN(response, " ", 2)
	if len(parts) < 2 {
		return 0, "UNKNOWN", response
	}

	_, err := fmt.Sscanf(parts[0], "%d", &statusCode)
	if err != nil {
		return 0, "UNKNOWN", response
	}

	rest := parts[1]
	colonIdx := strings.Index(rest, ": ")
	if colonIdx == -1 {
		return statusCode, rest, ""
	}

	statusText = rest[:colonIdx]
	message = rest[colonIdx+2:]

	return statusCode, statusText, message
}
