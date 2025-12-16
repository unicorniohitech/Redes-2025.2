# Redes-2025.2

## Descrição

Projeto de aplicação cliente-servidor UDP em Go. O servidor implementa um sistema de dicionário distribuído com comandos LOOKUP, INSERT e UPDATE. A comunicação utiliza um protocolo UDP customizado com gerenciamento de confiabilidade, fragmentação de mensagens, detecção de perda de pacotes e verificação de integridade com CRC.

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

#### Formato de Comunicação

**Pacote UDP (Cliente → Servidor / Servidor → Cliente):**

Cada pacote UDP contém um header estruturado (4 bytes) seguido pela payload com o comando/resposta. Para mensagens maiores que 1024 bytes, múltiplos pacotes são enviados com os mesmos campos de header em cada fragmento.

## Protocolo UDP Customizado

### Estrutura do Pacote

```text
Byte 0-1:    Packet Number (uint16)     - Número sequencial (índice do fragmento)
Byte 2-3:    Total Packets (uint16)     - Quantidade total de fragmentos
Byte 4+:     Payload (variável)         - Dados da mensagem
Byte N-N+1:  Checksum CRC16 (uint16)    - Verificação de integridade
```

**Header Size**: 4 bytes  
**Max Payload**: 1024 bytes  
**Max Total Packet**: ~1030 bytes

#### Campos do Header

| Campo | Bytes | Tipo | Descrição |
|-------|-------|------|-----------|
| **Packet Number** | 0-1 | uint16 | Índice sequencial do pacote dentro de um fragmento (0-based) |
| **Total Packets** | 2-3 | uint16 | Quantidade total de pacotes para reassembly completo |
| **Payload** | 4+ | []byte | Dados da mensagem (comando/resposta), até 1024 bytes |
| **CRC16** | N-N+1 | uint16 | Checksum calculado sobre todo o pacote exceto este campo |

### Verificação de Integridade

O protocolo implementa verificação robusta de integridade utilizando CRC16:

#### Campos de Verificação

| Campo | Descrição |
|-------|-----------|
| **Packet Number** | Identifica a ordem sequencial do pacote (índice 0-based) dentro de um fragmento |
| **Total Packets** | Quantidade total de fragmentos esperados para reassembly completo |
| **CRC16** | Checksum de 16 bits calculado sobre todo o pacote (header + payload) |

#### Algoritmo CRC

- **Tipo**: CRC-16 com polynomial 0x1021
- **Escopo**: Calcula sobre todos os bytes do pacote (Packet Number + Total Packets + Payload)
- **Cálculo**: Campo CRC zerado durante cálculo, preenchido após computação
- **Detecção**: Identifica erros de transmissão, corrupção de bits e pacotes malformados
- **Aceitação**: Pacote rejeitado se CRC não corresponder

#### Reassembly de Fragmentos

1. Pacotes fragmentados são identificados por `Total Packets > 1`
2. Validação individual: cada fragmento é verificado pelo CRC16
3. Ordem de reconstrução: usa `Packet Number` (0-based) para ordenação sequencial
4. Detecção de completude: verifica se `len(packets) == Total Packets`
5. Reassembly: concatena payloads dos fragmentos na ordem correta
6. Descarte: qualquer fragmento com CRC inválido causa descarte de todo o lote

### Tipos de Mensagem

O protocolo UDP utiliza a estratégia de fragmentação para diferenciar tipos de comunicação através dos campos `Packet Number` e `Total Packets`:

| Tipo | Identificação | Propósito |
|------|---------------|-----------|
| REQUEST | Single packet (Total Packets = 1) | Comando do cliente |
| RESPONSE | Single/Multiple packets | Resposta do servidor |
| ACK | Single packet (Total Packets = 1) | Confirmação de recebimento |

#### Respostas UDP

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
- Reassembly automático com validação de integridade
- Ordem mantida via `Packet Number` sequencial

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

## Estrutura do Projeto

```bash
udp/
├── main.go           # Ponto de entrada da aplicação
├── go.mod            # Gerenciamento de dependências
├── server/
│   ├── server.go     # Lógica do servidor
│   ├── config.go     # Configuração do servidor
│   ├── db.go         # Banco de dados em memória
│   └── utils.go      # Funções auxiliares do servidor
├── client/
│   ├── client.go     # Lógica do cliente
│   ├── config.go     # Configuração do cliente
│   ├── test.go       # Funções de teste
│   └── utils.go      # Funções auxiliares do cliente
├── utils/
│   ├── packet.go     # Estrutura e manipulação de pacotes
│   ├── crc.go        # Cálculo de CRC16
│   ├── http.go       # Utilitários HTTP
│   └── logger.go     # Sistema de logging
└── test_files/
    ├── golang.txt    # Arquivo com mais de 2000 bytes em texto plano
    └── python.txt    # Arquivo com mais de 2000 bytes em texto plano
```
