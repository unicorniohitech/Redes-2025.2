# Redes-2025.2

## Descrição

Projeto de aplicação cliente-servidor HTTP REST em Go. O servidor implementa um sistema de dicionário distribuído com comandos LOOKUP, INSERT e UPDATE. A comunicação utiliza estruturas `HTTPRequest` e `HTTPResponse` personalizadas, com códigos de status HTTP apropriados para operações de dicionário distribuído.

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
   cd Redes-2025.2/http-rest
   ```

2. Instale as dependências:

   ```bash
   go mod download
   ```

## Como Executar

### Passo 1: Iniciar o Servidor

Em um terminal, navegue até o diretório `http-rest` e execute:

```bash
cd http-rest
go run main.go -mode=server
```

O servidor iniciará na porta padrão `8000` no endereço `localhost`. Para especificar uma porta ou endereço diferentes:

```bash
go run main.go -mode=server -address=localhost -port=8000
```

### Passo 2: Iniciar o Cliente

Em outro terminal, navegue até o diretório `http-rest` e execute:

```bash
cd http-rest
go run main.go -mode=client
```

O cliente se conectará ao servidor em `localhost:8000`. Para conectar a um servidor diferente:

```bash
go run main.go -mode=client -address=localhost -port=8000
```

### Uso do Cliente

Após conectar, digite comandos no terminal do cliente:

#### Comandos Disponíveis

- **`LIST`** - Lista todos os termos cadastrados
- **`LOOKUP <termo>`** - Consulta a definição de um termo
- **`INSERT <termo> <definição>`** - Insere um novo termo no dicionário
- **`UPDATE <termo> <nova_definição>`** - Atualiza a definição de um termo existente

#### Formato de Comunicação

**HTTPRequest (Cliente → Servidor):**

Cada requisição HTTP contém um método (INSERT, LOOKUP, UPDATE ou LIST), um path com o termo e, quando aplicável, um corpo com a definição.

```bash
METHOD /term
Body: definition (quando aplicável)
```

**HTTPResponse (Servidor → Cliente):**

Todas as respostas HTTP seguem o padrão:

```bash
<StatusCode> <StatusText>: <Message>
```

#### Respostas HTTP

Todas as respostas seguem o formato: `<StatusCode> <StatusText>: <Message>`

**Códigos de Status:**

- `200 OK` - Operação bem-sucedida (LOOKUP, UPDATE)
- `201 Created` - Termo inserido com sucesso
- `400 Bad Request` - Formato de comando inválido
- `404 Not Found` - Termo não encontrado
- `408 Request Timeout` - Timeout ao acessar o dicionário
- `409 Conflict` - Termo já existe (INSERT)
- `501 Not Implemented` - Comando desconhecido

Para encerrar, pressione `Ctrl+C`

## Parâmetros de Linha de Comando

- `-mode`: **obrigatório** - Define o modo de execução (`server` ou `client`)
- `-address`: opcional - Endereço para bind/conexão (padrão: `localhost`)
- `-port`: opcional - Porta para bind/conexão (padrão: `8000`)

## Exemplo de Uso

**Terminal 1 (Servidor):**

```bash
cd http-rest
go run main.go -mode=server -port=8080
```

**Terminal 2 (Cliente):**

```bash
cd http-rest
go run main.go -mode=client -port=8080
```

**Interação:**

```bash
Enter message to send (or Ctrl+C to quit):
INSERT golang A programming language
# Cliente envia: INSERT /golang\r\nBody: A programming language\r\n\r\n
# Servidor responde: 201 Created: Term 'golang' inserted successfully

LOOKUP golang
# Cliente envia: LOOKUP /golang\r\n\r\n
# Servidor responde: 200 OK: A programming language

UPDATE golang A statically typed programming language
# Cliente envia: UPDATE /golang\r\nBody: A statically typed programming language\r\n\r\n
# Servidor responde: 200 OK: Term 'golang' updated successfully

LOOKUP python
# Cliente envia: LOOKUP /python\r\n\r\n
# Servidor responde: 404 Not Found: Term 'python' not found

LIST
# Cliente envia: LIST
# Servidor responde: 200 OK: [golang]
```

## Estrutura do Projeto

```bash
http-rest/
├── main.go           # Ponto de entrada da aplicação
├── go.mod            # Gerenciamento de dependências
├── Dockerfile        # Configuração Docker para containerização
├── server/
│   ├── server.go     # Lógica do servidor HTTP REST
│   ├── config.go     # Configuração do servidor
│   ├── db.go         # Banco de dados em memória
│   └── utils.go      # Funções auxiliares do servidor
├── client/
│   ├── client.go     # Lógica do cliente HTTP
│   ├── config.go     # Configuração do cliente
│   └── utils.go      # Funções auxiliares do cliente
└── utils/
    ├── http.go       # Utilitários HTTP e estruturas de requisição/resposta
    └── logger.go     # Sistema de logging
```
