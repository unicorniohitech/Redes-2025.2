package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

type APIResponse struct {
	Success bool        `json:"sucesso"`
	Message string      `json:"mensagem,omitempty"`
	Data    interface{} `json:"dados,omitempty"`
}

func StartClient(config *Config) error {
	baseURL := "http://" + config.AddressString()

	for {
		menu := promptui.Select{
			Label: "Selecione um comando",
			Items: []string{"LISTAR", "BUSCAR", "INSERIR", "ATUALIZAR"},
		}

		_, command, err := menu.Run()
		if err != nil {
			return err
		}

		switch command {

		case "LISTAR":
			resp, err := http.Get(baseURL + "/termos")
			printResponse(resp, err)

		case "BUSCAR":
			term := readInput("Digite o termo")
			resp, err := http.Get(baseURL + "/termos/buscar?termo=" + term)
			printResponse(resp, err)

		case "INSERIR":
			term := readInput("Digite o termo")
			definition := readInput("Digite a definição")

			body, _ := json.Marshal(map[string]string{
				"termo":     term,
				"definicao": definition,
			})

			resp, err := http.Post(
				baseURL+"/termos/inserir",
				"application/json",
				bytes.NewBuffer(body),
			)
			printResponse(resp, err)

		case "ATUALIZAR":
			term := readInput("Digite o termo")
			definition := readInput("Digite a nova definição")

			body, _ := json.Marshal(map[string]string{
				"termo":     term,
				"definicao": definition,
			})

			req, _ := http.NewRequest(
				http.MethodPut,
				baseURL+"/termos/atualizar",
				bytes.NewBuffer(body),
			)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			printResponse(resp, err)
		}
	}
}

func printResponse(resp *http.Response, err error) {
	if err != nil {
		fmt.Println("Erro de conexão:", err)
		return
	}
	defer resp.Body.Close()

	var response APIResponse
	json.NewDecoder(resp.Body).Decode(&response)

	fmt.Println("\nStatus:", resp.Status)

	if response.Message != "" {
		fmt.Println(response.Message)
	}

	if response.Data != nil {
		switch data := response.Data.(type) {

		case []interface{}:
			if len(data) == 0 {
				fmt.Println("Sem termos cadastrados")
				break
			}
			fmt.Println("Dados:")
			for _, v := range data {
				fmt.Println(" -", v)
			}

		case map[string]interface{}:
			fmt.Println("Dados:")
			for k, v := range data {
				fmt.Printf("   %s: %v\n", k, v)
			}
		}
	}

	fmt.Println()
}

func readInput(label string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(label + ": ")
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}
