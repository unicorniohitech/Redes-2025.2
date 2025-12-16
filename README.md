# 🌐 Redes 2025.2

Repositório de projetos desenvolvidos para a disciplina **Redes de Computadores** (Período: 2025.2).

## 📌 Visão Geral

Este repositório reúne implementações de sistemas cliente-servidor que simulam comunicação em rede utilizando diferentes protocolos de transporte, como TCP e UDP. Os projetos têm como objetivo exercitar o entendimento prático dos conceitos de redes, incluindo:

- Estabelecimento de conexão
- Envio e recebimento de mensagens
- Controle de fluxo
- Confiabilidade x desempenho

Cada projeto é modular e pode ser executado de forma independente, facilitando estudos e apresentação dos resultados.

## 🚀 Tecnologias Utilizadas

### Linguagens e Ferramentas

- **Go**: linguagem utilizada no desenvolvimento do cliente e servidor TCP  
  (versão recomendada: `v1.25.4` ou superior)
- **Terminal / CLI**: execução dos programas via linha de comando
- **PromptUI** (no cliente TCP): interação mais intuitiva na interface do cliente
- **git / GitHub**: versionamento e organização do código

### Protocolos Implementados

- **TCP (Transmission Control Protocol)**:
  - Comunicação orientada à conexão
  - Entrega confiável de dados
- **UDP (User Datagram Protocol)**:
  - Comunicação sem conexão
  - Alta velocidade com menor overhead (em desenvolvimento)

## 📁 Projetos

### 📡 [TCP](./tcp)

- Implementação cliente-servidor com troca de mensagens
- Interface CLI interativa
- Suporte para diferentes modos de operação

🔗 **[Ver instruções detalhadas →](./tcp/README.md)**

---

### 📦 [UDP](./udp)

- Implementação cliente-servidor com troca de mensagens
- Interface CLI interativa
- Suporte para grandes payloads
- Suporte para diferentes modos de operação

🔗 **[Ver instruções detalhadas →](./udp/README.md)**

---

### 📦 [HTTP](./http-rest)

- Implementação cliente-servidor com troca de mensagens
- Interface CLI interativa
- Suporte para diferentes modos de operação

🔗 **[Ver instruções detalhadas →](./http-rest/README.md)**

---

## 🗂️ Estrutura do Repositório

```txt
Redes-2025.2/
├── http-rest/        # Projeto HTTP
│   ├── main.go
│   ├── server/
│   ├── client/
│   └── README.md     # Instruções HTTP
├── tcp/              # Projeto TCP
│   ├── main.go
│   ├── server/
│   ├── client/
│   └── README.md     # Instruções TCP
├── udp/              # Projeto UDP
│   ├── main.go
│   ├── server/
│   ├── client/
│   ├── test_files/   # Arquivos em texto plano com mais de 2000 bytes para teste
│   └── README.md     # Instruções UDP
└── README.md         # Este arquivo
```

## 🛠️ Como Executar o Projeto TCP

> 🔧 Pré-requisito: Docker (para o servidor) e Go (para o cliente quando executado pelos scripts)

1. Clone o repositório (se ainda não fez):

   ```bash
   git clone https://github.com/unicorniohitech/Redes-2025.2.git
   cd Redes-2025.2
   ```

2. Iniciar o servidor via Docker Compose (recomendado)
   - Entre na pasta `compose` e suba o serviço do servidor:

   ```bash
   cd compose
   docker compose up --build -d
   ```

   - Os servidores ficarão disponíveis em `localhost` nas portas `:8000`, `:8080` e `:9000` (conforme configuração do `compose/docker-compose.yaml`).
   - Para parar o servidor:

   ```bash
   docker compose down
   ```

3. Iniciar o cliente usando os scripts fornecidos (sem Docker) (apenas serviço tcp)
   - Os scripts estão em `client/` e aceitam dois parâmetros opcionais: `HOST` e `PORT` (valores padrão: `localhost` e `8000`).

   - Linux / macOS / WSL (Bash):

   ```bash
   cd client
   ./run_client.sh            # usa localhost:8000
   ./run_client.sh 127.0.0.1 8000
   ```

   - Windows (cmd / PowerShell):

   ```cmd
   cd client
   run_client.bat             # usa localhost:8000
   run_client.bat 127.0.0.1 8000
   ```

   - O script tenta executar o binário `tcp/bin/tcp` se existir; caso contrário, ele compila o projeto (`go build`) para `tcp/bin/tcp` e então executa o cliente.
   - Por isso, o script requer o `go` disponível no PATH para compilar na primeira execução.

4. Alternativa: executar direto com Go (menos propenso a erros de ambiente windows)
   - Se preferir executar sem os scripts, use diretamente o comando `go run` no módulo desejado:

   ```bash
   cd tcp
   go run main.go -mode=server           # servidor
   go run main.go -mode=client           # cliente (ou use os scripts)
   ```

   ou

   ```bash
   cd udp
   go run main.go -mode=server           # servidor
   go run main.go -mode=client           # cliente
   ```

   ou

   ```bash
   cd http-rest
   go run main.go -mode=server           # servidor
   go run main.go -mode=client           # cliente
   ```

5. Observações
   - Em sistemas Windows/macOS com Docker Desktop, se você executar clientes em contêineres, pode ser necessário usar `host.docker.internal` para alcançar o `localhost` do host; os scripts locais lidam com execução direta no host e não dependem de mapeamentos de rede do Docker.
