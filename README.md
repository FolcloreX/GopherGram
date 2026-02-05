# 🐹 GopherGram Uploader

**GopherGram** é uma ferramenta de automação de alta performance escrita em **Go (Golang)**, projetada para fazer upload de mídias inteiras, vídeos ou grandes volumes de arquivos para o Telegram.

Ele atua como um **Userbot** (cliente MTProto), permitindo uploads de até **2GB (ou 4GB para Premium)**, gerenciamento de canais e formatação automática de conteúdo.

---

## ✨ Funcionalidades Principais

- **🚀 Upload Resiliente:** Sistema de **Resume** automático. Se a internet cair ou o pc desligar, ele continua exatamente do arquivo onde parou (baseado no nome da pasta).
- **✂️ Split Inteligente:** Divide automaticamente vídeos e arquivos ZIP maiores que **2GB** (limite do Telegram) sem corromper o original.
- **🎥 Streaming & Preview:** Gera thumbnails e metadados (duração/resolução) via **FFmpeg** para que os vídeos toquem nativamente no player do Telegram.
- **🗂️ Organização Automática:**
  - Compacta arquivos de apoio (PDFs, Códigos) em ZIPs.
  - Envia vídeos na ordem correta dos módulos.
  - Gera um **Índice Navegável** (Menu) com hashtags (#F001, #F002...).
- **🤖 Automação de Infraestrutura:**
  - Se nenhum Chat ID for informado, **cria um canal novo** automaticamente com o nome do curso.
  - Atualiza a **Foto** e a **Descrição** do canal com estatísticas (Tamanho Total, Duração).
  - Gera link de convite.
- **📢 Divulgação:** Posta um Card final formatado em um Grupo/Tópico de "Feed" configurável.
- **🔐 Multi-Conta:** Suporta múltiplas sessões baseadas no número de telefone.

---

## 🛠️ Pré-requisitos

Antes de rodar, certifique-se de ter instalado:

1.  **Go 1.20+**: [Download Go](https://go.dev/dl/)
2.  **FFmpeg**: Essencial para processar vídeos e gerar thumbnails.
    - _Linux:_ `sudo apt install ffmpeg`
    - _Windows:_ [Baixar executável](https://ffmpeg.org/download.html) e adicionar ao PATH.
3.  **Credenciais do Telegram**: Obtenha seu `API_ID` e `API_HASH` em [my.telegram.org](https://my.telegram.org).

---

## ⚙️ Configuração (.env)

Crie um arquivo `.env` na raiz do projeto:

```env
# --- Credenciais da Conta (Obrigatório) ---
API_ID=123456
API_HASH=sua_hash_aqui
PHONE_NUMBER=+5511999999999
PASSWORD=sua_senha_2fa_se_tiver

# --- Configuração de Upload (Opcional) ---
# Se deixar vazio ou 0, o bot CRIA um CANAL NOVO com o nome da pasta.
# Se preencher, ele usa esse canal existente.
ORIGIN_CHAT_ID=

# --- Configuração de Divulgação (Opcional) ---
# ID do Grupo onde o Card Final será postado.
# Se vazio, envia para o seu "Saved Messages".
POST_GROUP_ID=-100123456789

# Se o grupo acima tiver tópicos, coloque o ID do tópico aqui.
POST_GROUP_TOPIC_ID=

# --- Personalização ---
# Assinatura que aparece no rodapé das mensagens
LOGO="Postado por @GopherGram"
```

---

## 🚀 Como Usar

O comando básico exige o caminho da pasta do curso.

### 1. Upload Simples (Capa Texto)

**Linux / macOS**

```bash
go run cmd/bot/main.go "/Caminho/Para/A/Midia"
go run cmd/bot/main.go "/caminho/para/midia" "/caminho/para/capa.jpg"

```

**Windows**

```bash
go run cmd\bot\main.go "C:\Caminho\Para\Midia"
```

### 2. Upload com Capa (Imagem)

Passe o caminho da imagem como segundo argumento. Ela será usada como foto do canal e no card de divulgação.

**Linux / macOS**

```bash
go run cmd/bot/main.go "/Caminho/Para/A/Midia" "/Caminho/Para/Capa.jpg"
```

**Windows**

```bash
go run cmd\bot\main.go "C:\Caminho\Para\Midia" "C:\Caminho\Para\Capa.jpg"
```

---

## 📂 Estrutura de Pastas Recomendada

Para garantir que a ordem dos vídeos fique correta (1, 2, 3...), numere suas pastas e arquivos:

```text
/Meu Curso de Golang
├── 01. Introdução
│   ├── 01. Instalação.mp4
│   ├── 02. Hello World.mp4
│   └── apostila.pdf  <-- Será zipado automaticamente
├── 02. Sintaxe Básica
│   ├── 01. Variáveis.mp4
│   └── 02. Funções.mp4
└── capa.jpg
```

---

## 🧠 Como funciona o Estado (Resume)

O bot cria uma pasta `session/` na raiz.

- **`session_+55...json`**: Guarda sua sessão de login (para não pedir código toda vez).
- **`progress_Nome_Do_Curso.json`**: Guarda quais arquivos já foram enviados e qual o ID do canal criado.

**Para reiniciar um upload do zero:** Basta apagar o arquivo `.json` referente àquele curso dentro da pasta `session/`.

## 📝 Licença

Este projeto está sob a licença MIT. Sinta-se livre para contribuir! 🤝
