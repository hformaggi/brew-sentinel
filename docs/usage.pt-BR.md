# Guia de uso

[Voltar ao README](../README.pt-BR.md) · [English](usage.md)

## Idioma

A interface segue o idioma do sistema: português em locales `pt`, inglês nos demais.
Inglês é o fallback. Em **Ferramentas → Idioma**, escolha Padrão do sistema, English ou
Português. A escolha vale imediatamente e fica salva no diretório de configuração
pessoal do sistema (`~/Library/Application Support/brew-sentinel/settings.json` no macOS).

Para mudar somente uma execução, use `--lang en`, `--lang pt` ou `--lang system`.
A detecção verifica `LC_ALL`, `LC_MESSAGES`, `LANGUAGE` e `LANG`; na ausência deles,
consulta os idiomas preferidos do macOS. Descrições e avisos externos mantêm o idioma
fornecido pelo Brew e pelos projetos de origem.

## Interface e navegação

A tela principal tem um resumo compacto no topo, uma barra lateral com a biblioteca
e as ferramentas, e a lista de pacotes. O resumo lateral do pacote aparece quando
há pelo menos 140 colunas; em telas menores, os detalhes abrem em uma página própria.
A análise de vulnerabilidades é automática após carregar o inventário; a medição de
disco começa em paralelo. Ambas permitem continuar navegando.

Todas as telas usam o mesmo padrão:

| Tecla | Ação |
|---|---|
| ↑ / ↓ | Percorrer itens ou rolar o texto |
| ← / → | Trocar área, aba ou botão |
| Enter | Abrir o item ou acionar o botão selecionado |
| Esc | Voltar ou cancelar |

Na lista, **←** leva ao menu lateral e **↑**, no primeiro pacote, leva à busca.
No menu, escolha com ↑/↓ e abra com Enter. → retorna à lista; quando estiver
sobre uma categoria da biblioteca, também aplica o filtro selecionado. Buscar, instalar, filtrar, ordenar,
manutenção, relatório e ajuda estão acessíveis ali, sem decorar atalhos.

Nos detalhes, ←/→ percorrem **Resumo, Dependências, Dependentes, Segurança e Ações**.
↑/↓ percorrem o conteúdo da aba; Enter abre um pacote relacionado ou uma ação.
Na árvore, → expande/entra num ramo e ← recolhe/volta ao pai.

No catálogo, ↑/↓ alternam campo de busca, tipo de pacote e botões. ←/→ escolhem
Fórmulas/Aplicativos ou Buscar/Voltar. Em campos de texto, ←/→ editam o cursor.
Enter no campo de busca já inicia a pesquisa.

Nas prévias, ←/→ escolhem Cancelar/Confirmar; ↑ passa para o texto rolável e ←/→
retornam ao botão que já estava selecionado. Enter só aciona o botão em foco.
Nenhuma alteração acontece apenas por abrir ou ler uma prévia.

Atalhos opcionais: `/` busca, `i` catálogo, `m` manutenção, `s` relatório, `v` reanálise,
`r` recarregar, `o` ordenar, `?` ajuda e `q` sair. Ctrl+P abre a lista de comandos.
Ctrl+C interrompe consultas em andamento; operações mutáveis aguardam conclusão.

A busca aceita nome/descrição e filtros como `type:formula reason:explicit`,
`type:cask chrome` e `outdated:true`. O menu Filtrar lista também oferece opções
prontas para instalações explícitas, dependências e origem desconhecida.

## Como interpretar

- **Disponível** vem do metadata local do Brew, que pode estar desatualizado.
  Consultas da aplicação não executam `brew update` automaticamente.
- **Motivo** é lido do recibo quando disponível. Ausência de evidência é
  `unknown`; não significa que o pacote nunca é usado.
- **Dependências diretas** preferem `declared_directly` do recibo. Receita atual
  é fallback identificado; recibos incompletos preservam relações comprovadas.
  A árvore agregada não comprova uma árvore histórica completa de todos os kegs.
- **Runtime observado** pode conter relações transitivas. Não é apresentado
  como uma lista de relações diretas.
- **Disco** usa `du -sk`, sem seguir symlinks. Mede kegs, incluindo versões antigas,
  sem contabilizar uma dependência novamente para cada dependente. Valores são
  bytes alocados reportados pelo sistema, não espaço físico exclusivo em APFS.
  Casks ficam indisponíveis por não ser possível medir aplicativos externos ao
  Caskroom de maneira geral. Cache não está incluído nesse total.
- **Segurança** utiliza o scanner do Brew e pode consultar serviços externos.
  Casks não são suportados. O resultado informa cobertura, ignorados, falhas e
  patches; ausência de findings não garante ausência de vulnerabilidades.
  O scan não comprova todos os kegs antigos. Findings indicam versão upstream,
  que não deve ser confundida com revisão Brew. Uma atualização não é, por si só,
  comprovação de correção: recarregue e analise novamente.

## Operações

Instalação/atualização apresentam `--dry-run` antes de confirmar; remoção mostra
os dependentes conhecidos e deixa a validação final com o Brew. Limpeza e remoção
de órfãos são fluxos separados. Nunca são usados `--force`, `--zap` ou
`--ignore-dependencies`, nem exclusão manual de arquivos.

Operações de fórmulas mostram progresso em uma tela própria, com até 1 MiB de
saída em memória. Casks usam o terminal suspenso para preservar interação com
instaladores e senhas: a saída interativa não é capturada. A TUI é restaurada
após Enter e o inventário é recarregado, inclusive depois de falha parcial.
O estado pode mudar externamente entre prévia e execução.

Medições e análises atualizam os cartões e a lista sem abrir relatórios sobre a
navegação. Consultas de segurança e disco são executadas em paralelo; alterações e recargas do
inventário aguardam essas consultas terminarem. A tabela continua navegável.
