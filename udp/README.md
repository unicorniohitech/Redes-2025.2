# Redes-2025.2

## Descrição

Projeto de aplicação cliente-servidor UDP em Go. O servidor implementa um sistema de dicionário distribuído com comandos LOOKUP, INSERT e UPDATE. A comunicação utiliza um protocolo UDP customizado com gerenciamento de confiabilidade, fragmentação de mensagens e detecção de perda de pacotes.

## Requisitos

- **Go**: versão [1.25.4](https://go.dev/doc/install) ou superior.
  - *[Windows](https://go.dev/dl/go1.25.4.windows-amd64.msi)*
  - *[Linux](https://go.dev/dl/go1.25.4.linux-amd64.tar.gz)*
  - *[MacOS](https://go.dev/dl/)*
    - *[ARM64](https://go.dev/dl/go1.25.4.darwin-arm64.pkg)*
    - *[x86-64](https://go.dev/dl/go1.25.4.darwin-amd64.pkg)*
  - *[Source](https://go.dev/dl/go1.25.4.src.tar.gz)*
- **Dependências**: gerenciadas automaticamente pelo Go modules
  - `go.uber.org/zap` (logging)

## Instalação

1. Clone o repositório:

   ```bash
   git clone <url-do-repositorio>
   cd Redes-2025.2/udp
   ```

2. Instale as dependências:

   ```bash
   go mod download
   ```

## Como Executar

### Passo 1: Iniciar o Servidor

Em um terminal, navegue até o diretório `udp` e execute:

```bash
cd udp
go run main.go -mode=server
```

O servidor iniciará na porta padrão `8080` no endereço `localhost`. Para especificar uma porta ou endereço diferentes:

```bash
go run main.go -mode=server -address=localhost -port=8080
```

### Passo 2: Iniciar o Cliente

Em outro terminal, navegue até o diretório `udp` e execute:

```bash
cd udp
go run main.go -mode=client
```

O cliente se conectará ao servidor em `localhost:8080`. Para conectar a um servidor diferente:

```bash
go run main.go -mode=client -address=localhost -port=8080
```

### Uso do Cliente

Após conectar, digite comandos no terminal do cliente:

#### Comandos Disponíveis

- **`LIST`** - Lista todos os termos cadastrados
- **`LOOKUP <termo>`** - Consulta a definição de um termo
- **`INSERT <termo> <definição>`** - Insere um novo termo no dicionário
- **`UPDATE <termo> <nova_definição>`** - Atualiza a definição de um termo existente

## Protocolo UDP Customizado

### Estrutura do Pacote

```text
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
| HEARTBEAT | 3 | Keep-alive |

### Respostas UDP

Todas as respostas seguem o formato: `<StatusCode> <StatusText>: <Message>`

**Códigos de Status:**

- `200 OK` - Operação bem-sucedida (LOOKUP, UPDATE)
- `201 Created` - Termo inserido com sucesso
- `400 Bad Request` - Formato de comando inválido
- `404 Not Found` - Termo não encontrado
- `408 Request Timeout` - Timeout ao acessar o dicionário
- `409 Conflict` - Termo já existe (INSERT)
- `501 Not Implemented` - Comando desconhecido

## Gerenciamento de Confiabilidade

### ACK Tracking

- Cada pacote REQ recebe um ACK do servidor
- Timeouts detectam perdas e acionam retransmissão
- Máximo de retentativas: **3**

### Fragmentação

- Mensagens grandes são fragmentadas automaticamente
- Max payload por fragmento: 1024 bytes
- Reassembly automático no receptor

### Métricas Coletadas

- Total de pacotes enviados
- Total de pacotes recebidos
- Total de pacotes perdidos (detectados)
- Total de retransmissões
- Latência média (ms)
- Taxa de perda (%)

Para encerrar, pressione `Ctrl+C`

## Parâmetros de Linha de Comando

- `-mode`: **obrigatório** - Define o modo de execução (`server` ou `client`)
- `-address`: opcional - Endereço para bind/conexão (padrão: `localhost`)
- `-port`: opcional - Porta para bind/conexão (padrão: `8080`)

## Exemplo de Uso

**Terminal 1 (Servidor):**

```bash
cd udp
go run main.go -mode=server -port=8080
```

**Terminal 2 (Cliente):**

```bash
cd udp
go run main.go -mode=client -port=8080
```

**Interação:**

```bash
Enter message to send (or Ctrl+C to quit):
INSERT golang A programming language
# Servidor responde: 201 Created: Term 'golang' inserted successfully

LOOKUP golang
# Servidor responde: 200 OK: A programming language

UPDATE golang A statically typed programming language
# Servidor responde: 200 OK: Term 'golang' updated successfully

LOOKUP python
# Servidor responde: 404 Not Found: Term 'python' not found

LIST
# Servidor responde: 200 OK: golang
```

## Testes

Execute os testes com:

```bash
go test ./utils ./client ./server -v
```

Execute com cobertura:

```bash
go test -cover ./utils ./client ./server
```

Benchmark de performance:

```bash
go test -run='' -bench='.' -benchmem ./utils
```
