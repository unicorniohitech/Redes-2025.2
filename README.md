# Redes-2025.2

Repositório de projetos para a disciplina de Redes 2025.2.

## Projetos

Este repositório contém implementações de comunicação cliente-servidor utilizando diferentes protocolos de transporte.

### 📡 [TCP](./tcp)

Implementação de aplicação cliente-servidor utilizando protocolo TCP (Transmission Control Protocol).

- **Linguagem**: Go
- **Características**: Conexão confiável, controle de fluxo, garantia de entrega
- **Funcionalidade**: Cliente envia mensagens que são processadas e retornadas pelo servidor

**[📖 Ver instruções completas →](./tcp/README.md)**

### 📦 UDP

Implementação de aplicação cliente-servidor utilizando protocolo UDP (User Datagram Protocol).

- **Linguagem**: A definir
- **Características**: Comunicação sem conexão, baixa latência
- **Status**: Em desenvolvimento

**[📖 Ver instruções →](./udp/README.md)** *(em breve)*

## Estrutura do Repositório

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

## Como Usar

1. Navegue até a pasta do projeto desejado
2. Siga as instruções específicas no README.md de cada projeto
3. Execute o servidor e o cliente conforme documentado

## Requisitos Gerais

- **TCP**: Go 1.25.4 ou superior
- **UDP**: A definir

---

**Disciplina**: Redes de Computadores  
**Período**: 2025.2
