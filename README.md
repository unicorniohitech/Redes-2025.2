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

### 📦 UDP

- Projeto em desenvolvimento
- Comunicação leve e sem conexão
- Aguardando definição de requisitos

🔗 **[Ver instruções →](./udp/README.md)** *(em breve)*

---

## 🗂️ Estrutura do Repositório



```
Redes-2025.2/
├── tcp/              # Projeto TCP
│   ├── main.go
│   ├── server/
│   ├── client/
│   └── README.md     # Instruções TCP
├── udp/              # Projeto UDP
│   └── README.md     # Instruções UDP (em breve)
└── README.md         # Este arquivo
```

## 🛠️ Como Executar o Projeto TCP

> 🔧 Pré-requisito: Go 1.25.4 ou superior
1. Clone o repositório:
   ```bash
   git clone https://github.com/unicorniohitech/Redes-2025.2.git
   ```
2. Instale as dependências:
   ```bash
   go mod download
   ```
3. Acesse o diretório `tcp`:
   ```bash
   cd tcp
   ```
4. Para iniciar o servidor, execute:
   ```bash
   go run main.go -mode=server
   ```
5. Em outro terminal, inicie o cliente:
   ```bash
   go run main.go -mode=client
   ```
