# 📦 UDP - User Datagram Protocol Implementation

## 📌 Visão Geral

Implementação de um sistema cliente-servidor usando **UDP (User Datagram Protocol)** com gerenciamento robusto de pacotes, fragmentação de mensagens, detecção de perda e retransmissão automática.

Este projeto espelha a estrutura do projeto TCP, mas com as características específicas de um protocolo sem conexão e com confiabilidade implementada em camada de aplicação.

---

## 🎯 Objetivos

- **Comunicação sem conexão**: Usar UDP em vez de TCP
- **Confiabilidade manual**: Implementar ACKs e retransmissão
- **Gerenciamento de pacotes**: Fragmentação e reassembly
- **Detecção de perda**: Rastreamento e métricas
- **Compatibilidade**: Mesma interface de aplicação que TCP

---

## 🏗️ Estrutura do Projeto

```
udp/
├── main.go                 # Ponto de entrada
├── go.mod                  # Dependências Go
├── Dockerfile              # Build em container
├── README.md              # Este arquivo
│
├── server/
│   ├── server.go          # Listener UDP e handler
│   ├── db.go              # Dicionário em memória
│   ├── config.go          # Configuração do servidor
│   └── utils.go           # Utilitários do servidor
│
├── client/
│   ├── client.go          # Cliente UDP interativo
│   ├── config.go          # Configuração do cliente
│   └── utils.go           # Utilitários do cliente
│
├── utils/
│   ├── logger.go          # Logger (Zap)
│   ├── protocol.go        # Protocolo UDP customizado
│   ├── packet.go          # Gerenciamento de pacotes
│   └── reliability.go      # ACKs e retransmissão
│
└── bin/                    # Binários compilados
```

---

## 🔧 Protocolo UDP Customizado

### Estrutura do Pacote

```
Byte 0-3:    Packet ID (uint32)         - Identificador único
Byte 4:      Message Type (uint8)       - 0=REQ, 1=RES, 2=ACK, 3=HB
Byte 5-6:    Data Size (uint16)         - Tamanho da payload
Byte 7-8:    Total Packets (uint16)     - Pacotes no lote (fragmentação)
Byte 9-10:   Packet Number (uint16)     - Número deste pacote
Byte 11-12:  Checksum (uint16)          - CRC validação
Byte 13+:    Payload (variável)         - Comando/resposta
```

**Header Size**: 13 bytes  
**Max Payload**: 1024 bytes  
**Max Total Packet**: ~1050 bytes

### Tipos de Mensagem

| Tipo | Código | Propósito |
|------|--------|-----------|
| REQUEST | 0 | Comando do cliente |
| RESPONSE | 1 | Resposta do servidor |
| ACK | 2 | Confirmação de recebimento |
| HEARTBEAT | 3 | Keep-alive (futuro) |

---

## 🚀 Como Usar

### 🔨 Pré-requisitos

- Go 1.25.4 ou superior
- Acesso a terminal/PowerShell

### 💾 Instalação de Dependências

```bash
go mod download
```

### ▶️ Executar Servidor

```bash
# Modo desenvolvimento
go run main.go -mode=server -address=localhost -port=8000

# Ou compilar
go build -o bin/server .
./bin/server -mode=server -address=0.0.0.0 -port=8000
```

**Variáveis de Ambiente:**
```bash
HOST=0.0.0.0 PORT=9000 go run main.go -mode=server
```

### 👤 Executar Cliente

```bash
# Modo desenvolvimento
go run main.go -mode=client -address=localhost -port=8000

# Ou compilar
go build -o bin/client .
./bin/client -mode=client -address=localhost -port=8000
```

---

## 📡 Operações Disponíveis

### 1. **LIST**
Lista todos os termos no dicionário

```
Client → Server: "LIST"
Server → Client: "termo1\ntermo2\ntermo3\n..."
```

### 2. **LOOKUP <termo>**
Busca a definição de um termo

```
Client → Server: "LOOKUP termo"
Server → Client: "definição do termo"
```

### 3. **INSERT <termo> <definição>**
Insere um novo termo e sua definição

```
Client → Server: "INSERT novo definição"
Server → Client: "Success: termo inserido" ou "Error: termo já existe"
```

### 4. **UPDATE <termo> <nova_definição>**
Atualiza a definição de um termo existente

```
Client → Server: "UPDATE termo nova_definição"
Server → Client: "Success: termo atualizado" ou "Error: termo não existe"
```

---

## 🔄 Fluxo de Comunicação

### Requisição Bem-Sucedida

```
Cliente                          Servidor
  │
  ├─ [REQ] Pacote #1 ─────────→ │
  │                              │
  │                    [ACK] Pacote #1 ← ACK
  │                              │
  │                              │ Processa
  │                              │
  │              [RES] Pacote #1 ← Resposta
  │                              │
  └─ Exibe resultado
```

### Com Retransmissão (Timeout)

```
Cliente                          Servidor
  │
  ├─ [REQ] Pacote #1 ─────────→ │
  │                         (PERDIDO)
  │
  │ [Timeout: 2s]
  │
  ├─ [REQ] Pacote #1 (retry) ─→ │
  │                              │ Processa
  │                    [ACK] ← ACK
  │
  └─ Continua...
```

### Com Fragmentação

```
Cliente                          Servidor
  │
  ├─ [REQ] Pacote 1/3 ─────────→ │
  ├─ [REQ] Pacote 2/3 ─────────→ │
  ├─ [REQ] Pacote 3/3 ─────────→ │
  │                              │
  │              [ACK] Todos 1-3 ← Confirmação
  │                              │
  │                              │ Reassembly
  │                              │ Processa
  │
  │         [RES] 1 ou mais pak ← Resposta
  │
  └─ Exibe resultado
```

---

## 📊 Gerenciamento de Confiabilidade

### ACK Tracking
- Cada pacote REQ recebe um ACK do servidor
- Cada pacote RES deve ser confirmado pelo cliente (futuro)
- Timeouts detectam perdas e acionam retransmissão

### Retransmissão
- Timeout padrão: **2 segundos**
- Máximo de retentativas: **3**
- Exponential backoff: `timeout * (retryCount + 1)`

### Métricas Coletadas
- Total de pacotes enviados
- Total de pacotes recebidos
- Total de pacotes perdidos (detectados)
- Total de retransmissões
- Latência média (ms)
- Taxa de perda (%)

---

## 🧪 Testes e Simulação

### Simular Perda de Pacotes

Para testar confiabilidade, você pode simular perda no servidor ou cliente:

```go
// No config.go
config.SetSimulateLoss(true)    // Ativa simulação
config.SetLossRate(0.1)         // 10% de perda
```

Isso irá:
- Descartar aleatoriamente X% dos pacotes recebidos
- Forçar retransmissões automáticas
- Permitir observar comportamento de confiabilidade

### Exemplo de Teste

```bash
# Terminal 1: Servidor com simulação de 20% de perda
go run main.go -mode=server

# Terminal 2: Cliente enviando múltiplos comandos
go run main.go -mode=client

# Observe: Retransmissões e timeouts no console
```

---

## 🔐 Validação de Integridade

### Checksum
- Algoritmo: CRC16 ou simples (soma)
- Validado em cada pacote recebido
- Pacotes corrompidos são descartados

### Sequenciamento
- Cada pacote tem ID único (uint32)
- Detecta duplicatas
- Reassembly mantém ordem em fragmentos

---

## 📈 Performance

### Comparativo Esperado

| Métrica | TCP | UDP |
|---------|-----|-----|
| Latência | Maior (3-way handshake) | Menor |
| Confiabilidade | 100% | Configurável |
| Overhead | Maior (headers) | Menor |
| Complexidade | Simples | Complexa |
| Fragmentação | Automática | Manual |

---

## 🐳 Containerização

### Build da Imagem

```bash
docker build -t udp-app .
```

### Executar Container

```bash
# Servidor
docker run -p 8000:8000 -e MODE=server -e HOST=0.0.0.0 udp-app -mode=server

# Cliente (interativo)
docker run -it -e MODE=client -e HOST=host.docker.internal udp-app -mode=client
```

---

## 🛠️ Desenvolvimento

### Estrutura do Código

```
main.go
  ├─ Flag parsing
  └─ Mode selection
       ├─ Server Mode
       │  └─ server.StartServer()
       └─ Client Mode
          └─ client.StartClient()

server/
  ├─ server.go      → Listener + Handler
  ├─ config.go      → Configurações
  ├─ db.go          → Persistência de dados
  └─ utils.go       → Processamento de comandos

client/
  ├─ client.go      → Sender + Receiver
  ├─ config.go      → Configurações
  └─ utils.go       → Parser de respostas

utils/
  ├─ logger.go      → Logging
  ├─ protocol.go    → Serialização de pacotes
  ├─ packet.go      → Gerenciamento de buffers
  └─ reliability.go  → ACKs e métricas
```

### Adicionando Novos Comandos

1. Adicione comando em `client/utils.go`
2. Implemente handler em `server/utils.go`
3. Use `ProcessDictCommand()` como referência

---

## 🐛 Troubleshooting

### "Connection refused"
- Verifique se servidor está rodando
- Confirme porta e endereço
- Firewall pode estar bloqueando UDP

### "Timeout"
- Servidor não recebeu o pacote (perda)
- Resposta do servidor foi perdida
- Timeout é retentado automaticamente

### "Checksum failed"
- Pacote corrompido em trânsito
- Descartado automaticamente
- Cliente retransmite

---

## 📚 Recursos e Referências

- [RFC 768 - UDP Specification](https://tools.ietf.org/html/rfc768)
- [Go net package - UDPConn](https://golang.org/pkg/net/#UDPConn)
- [Uber Zap Logger](https://github.com/uber-go/zap)
- [PromptUI](https://github.com/manifoldco/promptui)

---

## 📝 Changelog

### v0.1.0 - Fundações
- [x] Estrutura base do projeto
- [x] Go.mod e configuração
- [x] Logger e utilitários
- [ ] Protocolo UDP customizado
- [ ] Gerenciamento de pacotes
- [ ] Servidor UDP
- [ ] Cliente UDP
- [ ] Testes

---

## 👥 Autor

Desenvolvido para a disciplina **Redes de Computadores 2025.2**

---

## 📄 Licença

Este projeto é fornecido como material educacional.
