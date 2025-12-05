# Corsarr CLI - Arquitetura e Planejamento

> 🏴‍☠️ Navigate the high seas of media automation

## 📋 Visão Geral

CLI em Golang para simplificar a configuração e inicialização da stack *arr (Radarr, Sonarr, etc). O usuário poderá selecionar serviços desejados, configurar variáveis de ambiente interativamente, e o CLI gerará automaticamente os arquivos `docker-compose.yml` e `.env` corretos.

### Problema Atual

- Múltiplos diretórios com `docker-compose.yml` diferentes (`vpn/`, `simple/`)
- Dificuldade de manutenção ao adicionar novos serviços
- Usuários precisam editar manualmente arquivos para escolher serviços
- Configuração manual de variáveis de ambiente propensa a erros

### Solução Proposta

CLI interativo que:

1. Permite seleção visual de serviços (checkboxes)
2. Configura variáveis de ambiente via prompts
3. Gera arquivos automaticamente baseado nas escolhas
4. Valida configurações antes de criar arquivos
5. Suporta profiles para reutilização de configurações

---

## 🏗️ Arquitetura do Projeto

```
corsarr-cli/
├── cmd/
│   ├── root.go           # Comando principal e configuração do Cobra
│   ├── generate.go       # Comando para gerar docker-compose e .env
│   ├── preview.go        # Preview das configurações antes de gerar
│   └── profile.go        # Gerenciar profiles salvos (save/load/list)
│
├── internal/
│   ├── i18n/
│   │   ├── i18n.go       # Sistema de internacionalização
│   │   ├── loader.go     # Carregamento de traduções
│   │   └── language.go   # Detecção e seleção de idioma
│   │
│   ├── services/
│   │   ├── services.go   # Definição de todos os serviços disponíveis
│   │   ├── categories.go # Categorização dos serviços
│   │   └── registry.go   # Registry pattern para gerenciar serviços
│   │
│   ├── generator/
│   │   ├── compose.go    # Geração do docker-compose.yml
│   │   ├── env.go        # Geração do arquivo .env
│   │   └── network.go    # Configuração de redes Docker
│   │
│   ├── validator/
│   │   ├── validator.go  # Validações de configuração
│   │   ├── ports.go      # Validação de conflitos de portas
│   │   └── dependencies.go # Validação de dependências entre serviços
│   │
│   ├── prompts/
│   │   ├── interactive.go # Prompts interativos (survey)
│   │   └── config.go      # Prompts de configuração de variáveis
│   │
│   └── profile/
│       ├── profile.go     # Estrutura de profiles
│       └── storage.go     # Persistência de profiles (YAML)
│
├── templates/
│   ├── docker-compose/
│   │   ├── base.tmpl            # Template base do compose (services, networks, volumes)
│   │   ├── vpn-mode.tmpl        # Configuração específica para modo VPN
│   │   └── network-mode.tmpl    # Configuração específica para modo network bridge
│   │
│   ├── services/                # Definições de cada serviço
│   │   ├── qbittorrent.yaml
│   │   ├── prowlarr.yaml
│   │   ├── flaresolverr.yaml
│   │   ├── sonarr.yaml
│   │   ├── radarr.yaml
│   │   ├── lidarr.yaml
│   │   ├── lazylibrarian.yaml
│   │   ├── bazarr.yaml
│   │   ├── jellyfin.yaml
│   │   ├── jellyseerr.yaml
│   │   ├── fileflows.yaml
│   │   └── gluetun.yaml
│   │
│   └── env.tmpl                 # Template do arquivo .env
│
├── locales/                     # Arquivos de tradução (i18n)
│   ├── en.yaml                  # English
│   ├── pt-br.yaml               # Português Brasileiro
│   └── es.yaml                  # Español
│
├── configs/
│   └── profiles/                # Diretório para profiles salvos
│
├── go.mod
├── go.sum
├── main.go
└── README.md
```

---

## 🔧 Serviços Identificados

### Download Managers
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| qBittorrent | 8081 | lscr.io/linuxserver/qbittorrent:latest | VPN, Simple |

### Indexers
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| Prowlarr | 9696 | lscr.io/linuxserver/prowlarr:latest | VPN, Simple |
| FlareSolverr | 8191 | ghcr.io/flaresolverr/flaresolverr:latest | VPN |

### Media Management
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| Sonarr (TV) | 8989 | lscr.io/linuxserver/sonarr:latest | VPN, Simple |
| Radarr (Movies) | 7878 | lscr.io/linuxserver/radarr:latest | VPN, Simple |
| Lidarr (Music) | 8686 | ghcr.io/hotio/lidarr:latest | Simple |
| LazyLibrarian (Books) | 5299 | lscr.io/linuxserver/lazylibrarian:latest | Simple |

### Subtitles
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| Bazarr | 6767 | ghcr.io/hotio/bazarr:latest | VPN, Simple |

### Streaming
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| Jellyfin | 8096 | lscr.io/linuxserver/jellyfin:latest | VPN, Simple |

### Request Management
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| Jellyseerr | 5055 | fallenbagel/jellyseerr:latest | VPN, Simple |

### Transcoding
| Serviço | Porta | Imagem | Presente em |
|---------|-------|--------|-------------|
| FileFlows | 19200 | revenz/fileflows:latest | VPN |

### VPN
| Serviço | Portas | Imagem | Presente em |
|---------|--------|--------|-------------|
| Gluetun | Múltiplas | qmcgaw/gluetun:latest | VPN |

---

## 📊 Estruturas de Dados

### Service
```go
type Service struct {
    ID            string              // Identificador único
    Name          string              // Nome amigável
    Category      ServiceCategory     // Categoria do serviço
    Image         string              // Imagem Docker
    ContainerName string              // Nome do container
    Hostname      string              // Hostname do container
    Ports         []PortMapping       // Mapeamento de portas
    Volumes       []VolumeMapping     // Mapeamento de volumes
    Environment   map[string]string   // Variáveis de ambiente específicas
    Devices       []string            // Dispositivos (ex: /dev/dri)
    RequiresVPN   bool                // Se requer VPN obrigatoriamente
    SupportsVPN   bool                // Se suporta VPN (opcional)
    Dependencies  []string            // IDs de serviços dependentes
    Optional      bool                // Se é opcional na configuração
    Description   string              // Descrição para o usuário
}

type ServiceCategory string

const (
    CategoryDownload    ServiceCategory = "Download Managers"
    CategoryIndexer     ServiceCategory = "Indexers"
    CategoryMedia       ServiceCategory = "Media Management"
    CategorySubtitles   ServiceCategory = "Subtitles"
    CategoryStreaming   ServiceCategory = "Streaming"
    CategoryRequest     ServiceCategory = "Request Management"
    CategoryTranscode   ServiceCategory = "Transcoding"
    CategoryVPN         ServiceCategory = "VPN"
)

type PortMapping struct {
    Host      string
    Container string
    Protocol  string // tcp, udp
}

type VolumeMapping struct {
    Host      string
    Container string
    ReadOnly  bool
}
```

### Configuration

```go
type Configuration struct {
    UseVPN       bool                // Se deve usar VPN
    Services     []string            // IDs dos serviços selecionados
    Environment  map[string]string   // Todas as variáveis de ambiente
    BasePath     string              // ARRPATH
    OutputDir    string              // Onde gerar os arquivos
    BackupOld    bool                // Se deve fazer backup dos arquivos antigos
}

type Profile struct {
    Name         string
    Description  string
    Configuration Configuration
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Environment Variables

```go
type EnvConfig struct {
    // Global
    ComposeProjectName string
    ARRPath            string
    Timezone           string
    PUID               string
    PGID               string
    UMASK              string
    
    // VPN (opcional)
    VPNServiceProvider  string
    VPNType             string
    WireguardPublicKey  string
    WireguardPrivateKey string
    WireguardAddresses  string
    VPNPortForwarding   string
    VPNDNSAddress       string
}
```

---

## 🌍 Sistema de Internacionalização (i18n)

### Idiomas Suportados
- 🇺🇸 **English (en)** - Padrão
- 🇧🇷 **Português Brasileiro (pt-br)**
- 🇪🇸 **Español (es)**

### Estrutura dos Arquivos de Tradução

Cada arquivo de locale (`locales/*.yaml`) contém todas as strings da interface:

```yaml
# locales/en.yaml
language:
  name: "English"
  code: "en"

prompts:
  language_select: "Select your language / Selecione seu idioma / Seleccione su idioma"
  vpn_question: "Do you want to use VPN (Gluetun)?"
  service_selection: "Select the services you want to use:"
  base_path: "Base path (ARRPATH):"
  timezone: "Timezone (TZ):"
  confirm_generation: "Confirm file generation?"
  save_profile: "Do you want to save this configuration as a profile?"
  profile_name: "Profile name:"

categories:
  download: "Download Managers"
  indexer: "Indexers"
  media: "Media Management"
  subtitles: "Subtitles"
  streaming: "Streaming"
  request: "Request Management"
  transcode: "Transcoding"
  vpn: "VPN"

services:
  qbittorrent:
    name: "qBittorrent"
    description: "BitTorrent client"
  radarr:
    name: "Radarr"
    description: "Movie collection manager"
  sonarr:
    name: "Sonarr"
    description: "TV show collection manager"
  # ... mais serviços

messages:
  validation_success: "Configuration validated successfully!"
  services_configured: "%d services will be configured"
  mode_vpn: "Mode: WITH VPN"
  mode_no_vpn: "Mode: WITHOUT VPN"
  no_port_conflicts: "No port conflicts detected"
  backup_created: "Backup created: %s"
  file_created: "File created successfully: %s"
  profile_saved: "Profile '%s' saved"

errors:
  invalid_path: "Invalid path: %s"
  port_conflict: "Port conflict detected: %d"
  missing_dependency: "Service '%s' requires '%s'"
  vpn_credentials_missing: "VPN credentials are missing"
```

```yaml
# locales/pt-br.yaml
language:
  name: "Português Brasileiro"
  code: "pt-br"

prompts:
  language_select: "Select your language / Selecione seu idioma / Seleccione su idioma"
  vpn_question: "Deseja usar VPN (Gluetun)?"
  service_selection: "Selecione os serviços que deseja usar:"
  base_path: "Caminho base (ARRPATH):"
  timezone: "Fuso horário (TZ):"
  confirm_generation: "Confirmar geração dos arquivos?"
  save_profile: "Deseja salvar esta configuração como perfil?"
  profile_name: "Nome do perfil:"

categories:
  download: "Gerenciadores de Download"
  indexer: "Indexadores"
  media: "Gerenciamento de Mídia"
  subtitles: "Legendas"
  streaming: "Streaming"
  request: "Gerenciamento de Requisições"
  transcode: "Transcodificação"
  vpn: "VPN"

# ... resto das traduções
```

```yaml
# locales/es.yaml
language:
  name: "Español"
  code: "es"

prompts:
  language_select: "Select your language / Selecione seu idioma / Seleccione su idioma"
  vpn_question: "¿Desea usar VPN (Gluetun)?"
  service_selection: "Seleccione los servicios que desea usar:"
  base_path: "Ruta base (ARRPATH):"
  timezone: "Zona horaria (TZ):"
  confirm_generation: "¿Confirmar generación de archivos?"
  save_profile: "¿Desea guardar esta configuración como perfil?"
  profile_name: "Nombre del perfil:"

categories:
  download: "Gestores de Descarga"
  indexer: "Indexadores"
  media: "Gestión de Medios"
  subtitles: "Subtítulos"
  streaming: "Streaming"
  request: "Gestión de Solicitudes"
  transcode: "Transcodificación"
  vpn: "VPN"

# ... resto das traduções
```

### Implementação do Sistema i18n

```go
// internal/i18n/i18n.go
package i18n

import (
    "embed"
    "fmt"
    
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var localeFS embed.FS

type I18n struct {
    bundle    *i18n.Bundle
    localizer *i18n.Localizer
    language  string
}

func New(lang string) (*I18n, error) {
    bundle := i18n.NewBundle(language.English)
    bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    
    // Carregar todos os idiomas
    for _, locale := range []string{"en", "pt-br", "es"} {
        bundle.MustLoadMessageFile(fmt.Sprintf("locales/%s.yaml", locale))
    }
    
    localizer := i18n.NewLocalizer(bundle, lang)
    
    return &I18n{
        bundle:    bundle,
        localizer: localizer,
        language:  lang,
    }, nil
}

func (i *I18n) T(key string, data ...interface{}) string {
    msg, err := i.localizer.Localize(&i18n.LocalizeConfig{
        MessageID: key,
        TemplateData: data,
    })
    if err != nil {
        return key // fallback para a chave se tradução não existir
    }
    return msg
}

func (i *I18n) GetLanguage() string {
    return i.language
}
```

```go
// internal/i18n/language.go
package i18n

import (
    "github.com/AlecAivazis/survey/v2"
)

type Language struct {
    Name string
    Code string
}

var SupportedLanguages = []Language{
    {Name: "🇺🇸 English", Code: "en"},
    {Name: "🇧🇷 Português Brasileiro", Code: "pt-br"},
    {Name: "🇪🇸 Español", Code: "es"},
}

func SelectLanguage() (string, error) {
    var selected string
    prompt := &survey.Select{
        Message: "Select your language / Selecione seu idioma / Seleccione su idioma:",
        Options: []string{
            SupportedLanguages[0].Name,
            SupportedLanguages[1].Name,
            SupportedLanguages[2].Name,
        },
        Default: SupportedLanguages[0].Name,
    }
    
    if err := survey.AskOne(prompt, &selected); err != nil {
        return "", err
    }
    
    // Mapear seleção para código
    for _, lang := range SupportedLanguages {
        if lang.Name == selected {
            return lang.Code, nil
        }
    }
    
    return "en", nil // fallback
}
```

### Uso no CLI

```go
// cmd/generate.go
package cmd

import (
    "github.com/spf13/cobra"
    "github.com/woliveiras/corsarr/internal/i18n"
    "github.com/woliveiras/corsarr/internal/prompts"
)

var generateCmd = &cobra.Command{
    Use:   "generate",
    Short: "Generate docker-compose.yml and .env",
    Run: func(cmd *cobra.Command, args []string) {
        // 1. Selecionar idioma PRIMEIRO
        langCode, err := i18n.SelectLanguage()
        if err != nil {
            panic(err)
        }
        
        // 2. Inicializar i18n com idioma selecionado
        translator, err := i18n.New(langCode)
        if err != nil {
            panic(err)
        }
        
        // 3. Usar tradutor em todo o fluxo
        vpnEnabled := prompts.AskVPN(translator)
        services := prompts.SelectServices(translator)
        config := prompts.ConfigureEnvironment(translator)
        
        // ... resto da lógica
    },
}
```

---

## 🎨 Fluxo de Uso

### 1. Modo Interativo Completo
```bash
./corsarr generate

# Passo 0: Seleção de Idioma (NOVO!)
? Select your language / Selecione seu idioma / Seleccione su idioma:
  > 🇺🇸 English
    🇧🇷 Português Brasileiro
    🇪🇸 Español

# === Se escolher Português Brasileiro ===

# Passo 1: Configuração de VPN
? Deseja usar VPN (Gluetun)? (s/N) › Não

# Passo 2: Seleção de Serviços
? Selecione os serviços que deseja usar:
  Download Managers:
    ☑ qBittorrent
  
  Indexers:
    ☑ Prowlarr
    ☐ FlareSolverr (requer VPN)
  
  Media Management:
    ☑ Sonarr (TV Shows)
    ☑ Radarr (Movies)
    ☐ Lidarr (Music)
    ☐ LazyLibrarian (Books)
  
  Subtitles:
    ☑ Bazarr
  
  Streaming:
    ☑ Jellyfin
  
  Request Management:
    ☐ Jellyseerr (requer Jellyfin)
  
  Transcoding:
    ☐ FileFlows

# Passo 3: Configuração Básica
? Caminho base (ARRPATH): › /home/chinelo/corsarr/
? Timezone (TZ): › Europe/Madrid
? User ID (PUID): › 1000
? Group ID (PGID): › 1000
? UMASK: › 002

# Passo 4: Validação
✓ Configuração validada com sucesso!
✓ 6 serviços serão configurados
✓ Modo: SEM VPN
✓ Nenhum conflito de portas detectado

# Passo 5: Preview
Preview dos arquivos que serão criados:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📄 docker-compose.yml (98 linhas)
📄 .env (8 variáveis)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

? Deseja salvar esta configuração como profile? (y/N) › Yes
? Nome do profile: › basico-sem-vpn

? Confirma a geração dos arquivos? (Y/n) › Yes

# Passo 6: Geração
✓ Backup criado: docker-compose.yml.backup
✓ Backup criado: .env.backup
✓ docker-compose.yml criado com sucesso
✓ .env criado com sucesso
✓ Profile 'basico-sem-vpn' salvo

Arquivos criados em: /home/chinelo/corsarr/

Para iniciar os serviços, execute:
  cd /home/chinelo/corsarr/
  docker compose up -d

Para verificar os logs:
  docker compose logs -f
```

### 2. Usando Profile Existente
```bash
./corsarr generate --profile basico-sem-vpn

✓ Profile 'basico-sem-vpn' carregado
✓ docker-compose.yml criado
✓ .env criado
```

### 3. Preview sem Gerar
```bash
./corsarr preview

# Mostra o conteúdo dos arquivos que seriam gerados
```

### 4. Modo Não-Interativo (CI/CD)
```bash
./corsarr generate --config config.yaml --no-interactive
```

---

## 🔍 Validações Implementadas

### 1. Conflitos de Portas
- Verifica se há portas duplicadas entre serviços
- Alerta sobre portas já em uso no sistema (opcional)

### 2. Dependências de Serviços
```
Jellyseerr → requer Jellyfin
FlareSolverr → útil com Prowlarr
Bazarr → requer Sonarr OU Radarr
FileFlows → requer Jellyfin
```

### 3. Validação de Paths
- Verifica se ARRPATH existe ou pode ser criado
- Valida permissões de escrita
- Verifica espaço disponível (aviso se < 10GB)

### 4. Validação de VPN
- Se VPN selecionado, valida credenciais obrigatórias
- Verifica formato de chaves Wireguard
- Valida provider suportado pelo Gluetun

### 5. Validação de Ambiente
- Verifica se Docker está instalado
- Verifica se Docker Compose está instalado
- Valida versão mínima do Docker

---

## 🎁 Features Adicionais

### 1. Sistema de Profiles
```bash
# Salvar configuração atual
./corsarr profile save completo

# Listar profiles
./corsarr profile list

# Carregar profile
./corsarr generate --profile completo

# Remover profile
./corsarr profile delete completo

# Exportar profile
./corsarr profile export completo > completo.yaml

# Importar profile
./corsarr profile import completo.yaml
```

### 2. Backup Automático
- Antes de gerar novos arquivos, faz backup dos existentes
- Formato: `docker-compose.yml.backup.TIMESTAMP`
- Mantém últimos 5 backups (configurável)

### 3. Modo Dry-Run
```bash
./corsarr generate --dry-run
# Apenas mostra o que seria feito, sem criar arquivos
```

### 4. Update de Serviços
```bash
./corsarr update
# Atualiza definições de serviços do repositório
```

### 5. Health Check
```bash
./corsarr health
# Verifica se todos os serviços configurados estão rodando
# Mostra status de cada container
```

### 6. Ports Check
```bash
./corsarr check-ports
# Verifica quais portas estão em uso no sistema
# Sugere portas alternativas se houver conflito
```

---

## 📦 Dependências Go

```go
require (
    github.com/spf13/cobra v1.8.0        // CLI framework
    github.com/spf13/viper v1.18.2       // Configuração
    github.com/AlecAivazis/survey/v2 v2.3.7 // Prompts interativos
    gopkg.in/yaml.v3 v3.0.1              // Parse YAML
    github.com/fatih/color v1.16.0       // Cores no terminal
    github.com/olekukonko/tablewriter v0.0.5 // Tabelas
    github.com/nicksnyder/go-i18n/v2 v2.4.0  // Internacionalização
    golang.org/x/text v0.14.0            // Suporte a linguagens
    text/template                         // Templates Go nativos
)
```

---

## 🚀 Roadmap de Implementação

### Fase 1: Estrutura Base
- [ ] Criar estrutura de diretórios
- [ ] Inicializar go.mod
- [ ] Configurar Cobra CLI
- [ ] Definir structs principais

### Fase 2: Sistema de Internacionalização (i18n)
- [ ] Criar estrutura de locales/
- [ ] Implementar sistema de i18n com go-i18n
- [ ] Criar arquivo de tradução en.yaml (English)
- [ ] Criar arquivo de tradução pt-br.yaml (Português Brasileiro)
- [ ] Criar arquivo de tradução es.yaml (Español)
- [ ] Implementar seleção de idioma no início do CLI
- [ ] Integrar traduções em todos os prompts e mensagens

### Fase 3: Definição de Serviços
- [ ] Mapear todos os serviços dos compose atuais
- [ ] Criar registry de serviços
- [ ] Definir categorias e dependências
- [ ] Documentar cada serviço em múltiplos idiomas

### Fase 4: Templates
- [ ] Criar template base do docker-compose
- [ ] Criar definições YAML de cada serviço
- [ ] Criar template de .env
- [ ] Implementar parser de service definitions
- [ ] Testar geração de templates com diferentes combinações

### Fase 5: Interface Interativa
- [ ] Implementar prompt de seleção de idioma (PRIMEIRO PASSO)
- [ ] Implementar prompt de seleção de VPN
- [ ] Implementar prompt de seleção de serviços
- [ ] Implementar prompt de configuração de variáveis
- [ ] Implementar validações inline
- [ ] Garantir que todas as mensagens sejam traduzidas

### Fase 6: Geradores
- [ ] Implementar gerador de docker-compose.yml
- [ ] Implementar gerador de .env
- [ ] Implementar sistema de backup
- [ ] Testar geração com diferentes combinações

### Fase 7: Validações
- [ ] Validação de portas (mensagens traduzidas)
- [ ] Validação de dependências (mensagens traduzidas)
- [ ] Validação de paths (mensagens traduzidas)
- [ ] Validação de VPN (mensagens traduzidas)
- [ ] Validação de ambiente Docker (mensagens traduzidas)

### Fase 8: Sistema de Profiles
- [ ] Implementar save/load de profiles
- [ ] Implementar list profiles
- [ ] Implementar delete profile
- [ ] Implementar export/import
- [ ] Salvar preferência de idioma no profile

### Fase 9: Features Extras
- [ ] Comando preview (traduzido)
- [ ] Comando health (traduzido)
- [ ] Comando check-ports (traduzido)
- [ ] Modo dry-run (traduzido)
- [ ] Modo não-interativo

### Fase 9: Documentação
- [ ] README do CLI em EN, PT-BR e ES
- [ ] Documentação de comandos (multilíngue)
- [ ] Exemplos de uso em múltiplos idiomas
- [ ] Troubleshooting guide (multilíngue)
- [ ] Atualizar README principal do repositório

### Fase 10: Testes
- [ ] Testes unitários para geradores
- [ ] Testes unitários para validadores
- [ ] Testes de i18n (todas as chaves traduzidas)
- [ ] Testes de integração
- [ ] Testes com diferentes combinações de serviços

---

## 📝 Notas de Implementação

### Sistema de Templates Modular

A geração do `docker-compose.yml` funciona de forma modular:

#### 1. Definições de Serviços (YAML)
Cada serviço tem um arquivo YAML em `templates/services/` com todas as suas configurações:

```yaml
# templates/services/radarr.yaml
id: radarr
name: Radarr
category: media
description: Movie collection manager
image: lscr.io/linuxserver/radarr:latest
container_name: radarr

ports:
  - host: "7878"
    container: "7878"
    protocol: tcp

volumes:
  - host: "${ARRPATH}config/radarr"
    container: "/config"
  - host: "${ARRPATH}backup/radarr"
    container: "/data/backup"
  - host: "${ARRPATH}data/movies"
    container: "/data/movies"
  - host: "${ARRPATH}data/downloads"
    container: "/downloads"

environment:
  - "TZ=${TZ}"
  - "PUID=${PUID}"
  - "PGID=${PGID}"
  - "UMASK=${UMASK}"

# Configurações específicas de rede
network:
  vpn_mode:
    network_mode: "service:gluetun"
  bridge_mode:
    hostname: radarr
    networks:
      - media

restart: unless-stopped
supports_vpn: true
dependencies: []
optional: false
```

#### 2. Template Base (Go Template)
O template base em `templates/docker-compose/base.tmpl` estrutura o compose:

```yaml
services:
{{- range .Services }}
  {{ .ContainerName }}:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    {{- if eq $.Mode "vpn" }}
    network_mode: "{{ .Network.VPNMode.NetworkMode }}"
    {{- else }}
    hostname: {{ .Network.BridgeMode.Hostname }}
    networks:
      {{- range .Network.BridgeMode.Networks }}
      - {{ . }}
      {{- end }}
    {{- end }}
    restart: {{ .Restart }}
    volumes:
      {{- range .Volumes }}
      - {{ .Host }}:{{ .Container }}{{ if .ReadOnly }}:ro{{ end }}
      {{- end }}
    {{- if and (ne $.Mode "vpn") (.Ports) }}
    ports:
      {{- range .Ports }}
      - "{{ .Host }}:{{ .Container }}{{ if ne .Protocol "tcp" }}/{{ .Protocol }}{{ end }}"
      {{- end }}
    {{- end }}
    {{- if .Environment }}
    environment:
      {{- range .Environment }}
      - {{ . }}
      {{- end }}
    {{- end }}
    {{- if .Devices }}
    devices:
      {{- range .Devices }}
      - {{ . }}
      {{- end }}
    {{- end }}
    env_file:
      - ./.env
{{- end }}

{{- if ne .Mode "vpn" }}
networks:
  media:
    driver: bridge
{{- end }}
```

#### 3. Fluxo de Geração

```go
// Pseudocódigo do processo de geração

func GenerateDockerCompose(selectedServices []string, useVPN bool) error {
    // 1. Carregar definições de serviços selecionados
    services := []Service{}
    for _, serviceID := range selectedServices {
        serviceConfig := LoadServiceDefinition(serviceID) // carrega YAML
        services = append(services, serviceConfig)
    }
    
    // 2. Adicionar Gluetun se VPN habilitado
    if useVPN {
        gluetun := LoadServiceDefinition("gluetun")
        services = prepend(services, gluetun) // Gluetun primeiro
    }
    
    // 3. Ajustar configurações baseado no modo
    mode := "bridge"
    if useVPN {
        mode = "vpn"
        // Remove portas dos serviços (ficam no Gluetun)
        // Ajusta network_mode de cada serviço
    }
    
    // 4. Gerar compose usando template
    tmpl := template.Must(template.ParseFiles("templates/docker-compose/base.tmpl"))
    data := struct {
        Services []Service
        Mode     string
    }{
        Services: services,
        Mode:     mode,
    }
    
    // 5. Executar template e salvar arquivo
    output := executeTemplate(tmpl, data)
    saveFile("docker-compose.yml", output)
    
    return nil
}
```

#### 4. Vantagens dessa Abordagem

✅ **Modularidade**: Cada serviço é independente e auto-contido  
✅ **Fácil Manutenção**: Atualizar um serviço não afeta outros  
✅ **Escalabilidade**: Adicionar novo serviço = criar 1 arquivo YAML  
✅ **Reusabilidade**: Mesma definição funciona para VPN e network bridge  
✅ **Validação**: YAML pode ser validado por schema  
✅ **Documentação**: Definição do serviço é auto-documentada  

#### 5. Exemplo de Uso

```bash
# CLI lê os arquivos YAML disponíveis
services := LoadAllServiceDefinitions("templates/services/")

# Mostra para o usuário escolher
selected := PromptUserToSelectServices(services)

# Gera compose baseado na seleção
GenerateDockerCompose(selected, useVPN)
```

### Network Mode
- **VPN Mode**: Todos os serviços usam `network_mode: "service:gluetun"`
- **Simple Mode**: Todos os serviços usam rede bridge customizada `networks: [media]`

### Volumes
Padrão de volumes por categoria:
- **Download**: `/downloads`
- **Media**: `/data/movies`, `/data/tvshows`, `/data/music`, `/data/books`
- **Config**: `/config`
- **Backup**: `/data/backup`

### Environment Variables
Variáveis globais aplicadas a todos os serviços:
- `TZ`, `PUID`, `PGID`, `UMASK`

Variáveis específicas gerenciadas por serviço.

### Restart Policy
Todos os serviços usam `restart: unless-stopped`

---

## 🔐 Segurança

- [ ] Nunca logar senhas ou chaves
- [ ] Arquivo .env com permissões 600
- [ ] Validar inputs do usuário
- [ ] Sanitizar paths
- [ ] Não executar comandos shell com input do usuário

---

## 📚 Referências

- [Docker Compose Specification](https://docs.docker.com/compose/compose-file/)
- [Gluetun Documentation](https://github.com/qdm12/gluetun-wiki)
- [LinuxServer.io Images](https://fleet.linuxserver.io/)
- [Cobra CLI](https://cobra.dev/)
- [Survey (Prompts)](https://github.com/AlecAivazis/survey)

---

## 📊 Status Atual

**Última atualização**: 2025-12-05

**Status**: 📋 Planejamento completo

**Próximo passo**: Iniciar Fase 1 - Estrutura Base
