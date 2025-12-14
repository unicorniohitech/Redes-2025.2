package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"go.uber.org/zap"
	"tcp/utils"
)

var (
	dictionary = NewDictionary()
	mutex      sync.Mutex
)

type APIResponse struct {
	Success  bool        `json:"sucesso"`
	Message  string      `json:"mensagem,omitempty"`
	Data     interface{} `json:"dados,omitempty"`
}

func StartServer(config *Config) error {
	logger := utils.GetLogger()

	mux := http.NewServeMux()

	mux.HandleFunc("/termos", listTerms)
	mux.HandleFunc("/termos/buscar", lookupTerm)
	mux.HandleFunc("/termos/inserir", insertTerm)
	mux.HandleFunc("/termos/atualizar", updateTerm)

	server := &http.Server{
		Addr:    config.AddressString(),
		Handler: mux,
	}

	logger.Info("Servidor HTTP REST iniciado",
		zap.String("endereco", config.AddressString()))

	return server.ListenAndServe()
}

func writeJSON(w http.ResponseWriter, status int, resp APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func listTerms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Método não permitido",
		})
		return
	}

	mutex.Lock()
	terms := dictionary.List()
	mutex.Unlock()

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    terms,
	})
}

func lookupTerm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Método não permitido",
		})
		return
	}

	term := strings.TrimSpace(r.URL.Query().Get("termo"))
	if term == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "O termo não pode ser vazio",
		})
		return
	}

	mutex.Lock()
	definition, ok := dictionary.LookUp(term)
	mutex.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Termo não encontrado",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]string{
			"termo":     term,
			"definicao": definition,
		},
	})
}

func insertTerm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Método não permitido",
		})
		return
	}

	var payload struct {
		Termo     string `json:"termo"`
		Definicao string `json:"definicao"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "JSON inválido",
		})
		return
	}

	term := strings.TrimSpace(payload.Termo)
	definition := strings.TrimSpace(payload.Definicao)

	if term == "" || definition == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "Termo e definição não podem ser vazios",
		})
		return
	}

	mutex.Lock()
	ok := dictionary.Insert(term, definition)
	mutex.Unlock()

	if !ok {
		writeJSON(w, http.StatusConflict, APIResponse{
			Success: false,
			Message: "O termo já existe",
		})
		return
	}

	writeJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "Termo inserido com sucesso",
	})
}

func updateTerm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Método não permitido",
		})
		return
	}

	var payload struct {
		Termo     string `json:"termo"`
		Definicao string `json:"definicao"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "JSON inválido",
		})
		return
	}

	term := strings.TrimSpace(payload.Termo)
	definition := strings.TrimSpace(payload.Definicao)

	mutex.Lock()
	ok := dictionary.Update(term, definition)
	mutex.Unlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: "Termo não encontrado",
		})
		return
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Definição atualizada com sucesso",
	})
}
