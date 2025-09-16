# dbfmini

[![Version](https://img.shields.io/github/v/tag/alberto255345/dbfmini?label=version&sort=semver)](https://github.com/alberto255345/dbfmini/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/alberto255345/dbfmini.svg)](https://pkg.go.dev/github.com/alberto255345/dbfmini)
[![Go Report Card](https://goreportcard.com/badge/github.com/alberto255345/dbfmini)](https://goreportcard.com/report/github.com/alberto255345/dbfmini)

Leitor **puro Go** para arquivos **dBASE/DBF** (sem dependências de bibliotecas DBF de terceiros).
Suporta campos: **C, N, F, Y, L, D, I, T, B**.

## Instalação

```bash
go get github.com/alberto255345/dbfmini@latest
```

## Uso rápido

```go
package main

import (
    "fmt"
    "log"

    "github.com/alberto255345/dbfmini"
)

func main() {
    db, err := dbfmini.Open("clientes.dbf", nil)
    if err != nil {
        log.Fatalf("abrindo DBF: %v", err)
    }

    for {
        registros, err := db.ReadRecords(200) // lote opcional
        if err != nil {
            log.Fatalf("lendo registros: %v", err)
        }
        if len(registros) == 0 {
            break // EOF
        }
        for _, rec := range registros {
            saldo, _ := rec["SALDO"].(float64)
            fmt.Printf("%v => saldo %.2f\n", rec["NOME"], saldo)
        }
    }
}
```

## Codificações e modos de leitura

- **Codificações**: por padrão usamos `ISO-8859-1`. Ajuste com `OpenOptions.Encoding`:
  - `Encoding.Default` define a página de códigos de todos os campos (`CP850`, `CP437`, `CP1252`, `ISO-8859-1`, `UTF-8`).
  - `Encoding.PerField` permite sobrescrever por nome de campo, ex.: `Encoding{Default: "CP850", PerField: map[string]string{"NOME": "CP1252"}}`.
- **Modos**: `ReadStrict` (padrão) valida versão, tipos e dados inconsistentes, retornando erro ao primeiro problema. `ReadLoose` é tolerante: ignora campos/linhas inválidos sempre que possível e aceita arquivos com formatos mistos.
- **Registros deletados**: defina `IncludeDeleted: true` nas opções para obter os registros marcados como excluídos (`rec["_deleted"] == true`).

## Limitações

- Campos **M (memo)** ainda não são lidos; o esqueleto existe, mas falta interpretar `.DBT/.FPT`.
- Biblioteca **somente leitura**: não gera nem altera arquivos DBF.
- Ainda não há suporte a tipos específicos do Visual FoxPro (ex.: `General`, `Variant`).

## Roadmap

- Suporte completo a campos memo (`.DBT`/`.FPT`).
- APIs de escrita/atualização de registros.
- Ampliação do suporte a outros tipos de campo e validações específicas por versão.
