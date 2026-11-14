# Brew Sentinel

**Seus pacotes Homebrew, dependências e relatórios de segurança em uma interface de terminal.**

O Brew Sentinel ajuda você a explorar fórmulas e casks instalados, entender por que
um pacote está na sua máquina e revisar alterações antes de executá-las pelo Homebrew.

[English](README.md) · [Primeiros passos](#primeiros-passos) · [Guia de uso](docs/usage.pt-BR.md) · [Relatar um problema](https://github.com/hformaggi/brew-sentinel/issues)

## Recursos

- **Inventário de pacotes** — consulte versões instaladas e disponíveis, pacotes
  desatualizados, versões fixadas e motivos de instalação.
- **Explorador de dependências** — navegue por dependências e dependentes, usando
  recibos de instalação como evidência quando disponíveis.
- **Relatórios de vulnerabilidades** — analise fórmulas com `brew vulns --json`,
  consultando vulnerabilidades, patches das fórmulas e lacunas de cobertura.
- **Gerenciamento de pacotes** — pesquise no catálogo, instale, atualize, desinstale,
  limpe versões antigas e remova dependências órfãs com prévia e confirmação.
- **Busca e filtros** — combine nomes, descrições, tipos e motivos de instalação
  para encontrar os pacotes que procura.

## Primeiros passos

### Requisitos

- macOS com [Homebrew](https://brew.sh/) instalado e `brew` disponível no `PATH`.
- Go **1.27 ou superior** para compilar o projeto.
- Git para clonar o repositório.
- Para relatórios de vulnerabilidades, uma versão do Homebrew com `brew vulns --json`
  e acesso ao serviço externo usado pelo scanner.

### Homebrew

```sh
brew install hformaggi/tap/brew-sentinel
```

### Compilar a partir do código

```sh
git clone https://github.com/hformaggi/brew-sentinel.git
cd brew-sentinel
go build -o bin/brew-sentinel ./cmd/brew-sentinel
./bin/brew-sentinel
```

## Uso

Navegue com as setas e pressione **Enter** para abrir um pacote ou ferramenta.
O menu lateral dá acesso à busca, aos filtros, à ordenação, à manutenção e aos relatórios.

| Tecla | Ação |
| --- | --- |
| ↑ / ↓ | Percorrer itens ou rolar texto |
| ← / → | Trocar área, aba ou botão |
| Enter | Abrir ou acionar a seleção |
| Esc | Voltar ou cancelar |
| `/` | Buscar pacotes |
| Ctrl+P | Abrir a lista de comandos |
| `?` | Mostrar ajuda |
| `q` | Sair |

### Exemplos de busca

Digite estas expressões no campo de busca de pacotes. Os filtros podem ser combinados.

| Consulta | Resultado |
| --- | --- |
| `type:formula reason:explicit` | Fórmulas instaladas explicitamente |
| `type:cask chrome` | Casks que correspondem a “chrome” |
| `outdated:true` | Pacotes com atualização nos metadados locais |
| `reason:dependency` | Pacotes instalados como dependência |

### Opções de linha de comando

```sh
./bin/brew-sentinel --lang en      # Usar inglês nesta execução
./bin/brew-sentinel --lang pt      # Usar português nesta execução
./bin/brew-sentinel --lang system  # Seguir o idioma do sistema
./bin/brew-sentinel --json         # Imprimir o inventário sem abrir a interface
./bin/brew-sentinel --version      # Mostrar a versão da compilação
./bin/brew-sentinel --help         # Listar as opções disponíveis
```

Consulte o [guia de uso](docs/usage.pt-BR.md) para ver todos os atalhos, a navegação
por dependências, o comportamento das operações e como interpretar os resultados.

## Alterações e relatórios

Navegar pelos pacotes não altera a instalação. Instalações e atualizações mostram
uma prévia do Homebrew antes da confirmação; a desinstalação mostra os dependentes
conhecidos e deixa a validação final com o Homebrew. Limpeza e remoção de órfãos têm
prévias próprias. A aplicação não força operações nem ignora verificações de dependências.

As versões disponíveis vêm dos metadados locais do Homebrew. O Brew Sentinel não
executa `brew update` automaticamente; atualize esses metadados quando necessário.

Os relatórios de vulnerabilidades cobrem fórmulas, sem cobertura confirmada de casks
ou de todos os kegs antigos. Um relatório vazio indica que não houve correspondências
conhecidas no escopo analisado. Verifique a cobertura e analise novamente após atualizar.
O scanner usa o serviço externo do Homebrew. Inventário e resultados de operações
ficam em memória, a menos que você salve a saída JSON.

## Desenvolvimento

```sh
go test ./...
go test -race ./...
go vet ./...
```

Os testes usam inventários sintéticos, comandos Homebrew simulados, diretórios
temporários e telas de terminal simuladas. Eles não instalam nem removem pacotes reais.

Para executar os benchmarks dos filtros:

```sh
go test ./internal/tui -bench . -run '^$'
```

Contribuições são bem-vindas por [issues](https://github.com/hformaggi/brew-sentinel/issues)
e [pull requests](https://github.com/hformaggi/brew-sentinel/pulls).
Ao relatar um problema, informe as versões do macOS, Homebrew e Brew Sentinel,
os passos para reproduzi-lo e as saídas relevantes, removendo informações pessoais.

## Licença

[BSD 3-Clause](LICENSE) © 2026 hformaggi.
