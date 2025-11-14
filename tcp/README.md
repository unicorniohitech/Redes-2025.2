# Redes-2025.2

## Descrição

Projeto de aplicação cliente-servidor TCP em Go. O servidor recebe mensagens dos clientes, processa os dados (convertendo para maiúsculas) e retorna a resposta. O cliente envia mensagens (convertidas para minúsculas) e exibe as respostas do servidor.

## Requisitos

- **Go**: versão 1.25.4 ou superior
- **Dependências**: gerenciadas automaticamente pelo Go modules
  - `go.uber.org/zap` (logging)

## Instalação

1. Clone o repositório:

   ```bash
   git clone <url-do-repositorio>
   cd Redes-2025.2/tcp
   ```

2. Instale as dependências:

   ```bash
   go mod download
   ```

## Como Executar

### Passo 1: Iniciar o Servidor

Em um terminal, navegue até o diretório `tcp` e execute:

```bash
cd tcp
go run main.go -mode=server
```

O servidor iniciará na porta padrão `8000` no endereço `localhost`. Para especificar uma porta ou endereço diferentes:

```bash
go run main.go -mode=server -address=localhost -port=9000
```

### Passo 2: Iniciar o Cliente

Em outro terminal, navegue até o diretório `tcp` e execute:

```bash
cd tcp
go run main.go -mode=client
```

O cliente se conectará ao servidor em `localhost:8000`. Para conectar a um servidor diferente:

```bash
go run main.go -mode=client -address=localhost -port=9000
```

### Uso do Cliente

Após conectar, digite mensagens no terminal do cliente:

1. Digite uma mensagem e pressione Enter
2. A mensagem será enviada ao servidor (convertida para minúsculas)
3. O servidor responderá com a mensagem em maiúsculas
4. Para encerrar, pressione `Ctrl+C`

## Parâmetros de Linha de Comando

- `-mode`: **obrigatório** - Define o modo de execução (`server` ou `client`)
- `-address`: opcional - Endereço para bind/conexão (padrão: `localhost`)
- `-port`: opcional - Porta para bind/conexão (padrão: `8000`)

## Exemplo de Uso

**Terminal 1 (Servidor):**

```bash
cd tcp
go run main.go -mode=server -port=8080
```

**Terminal 2 (Cliente):**

```bash
cd tcp
go run main.go -mode=client -port=8080
```

**Interação:**

```bash
Enter message to send (or Ctrl+C to quit):
Olá Mundo
# Cliente envia: "olá mundo"
# Servidor responde: "OLÁ MUNDO"
```

## Estrutura do Projeto

```bash
tcp/
├── main.go           # Ponto de entrada da aplicação
├── go.mod            # Gerenciamento de dependências
├── server/
│   ├── server.go     # Lógica do servidor
│   ├── config.go     # Configuração do servidor
│   └── utils.go      # Funções auxiliares do servidor
├── client/
│   ├── client.go     # Lógica do cliente
│   ├── config.go     # Configuração do cliente
│   └── utils.go      # Funções auxiliares do cliente
└── utils/
    └── logger.go     # Sistema de logging
```
