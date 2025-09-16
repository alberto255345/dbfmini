# dbfmini

[![Version](https://img.shields.io/github/v/tag/alberto255345/dbfmini?label=version&sort=semver)](https://github.com/alberto255345/dbfmini/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/alberto255345/dbfmini.svg)](https://pkg.go.dev/github.com/alberto255345/dbfmini)
[![Go Report Card](https://goreportcard.com/badge/github.com/alberto255345/dbfmini)](https://goreportcard.com/report/github.com/alberto255345/dbfmini)

Leitor **puro Go** para arquivos **dBASE/DBF** (sem dependências de bibliotecas DBF de terceiros).  
Suporta campos: **C, N, F, Y, L, D, I, T, B**.

## Instalação

```bash
go get github.com/alberto255345/dbfmini@latest
````

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
        registros, err := db.ReadRecords(200) // leitura em lotes (opcional)
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

* **Codificações**: por padrão `ISO-8859-1`. Ajuste com `OpenOptions.Encoding`:

  * `Encoding.Default` define a página de códigos de todos os campos (`CP850`, `CP437`, `CP1252`, `ISO-8859-1`, `UTF-8`).
  * `Encoding.PerField` permite sobrescrever por nome de campo, ex.:

    ```go
    Encoding{Default: "CP850", PerField: map[string]string{"NOME": "CP1252"}}
    ```
* **Modos**:

  * `ReadStrict` (padrão) valida versão/tipos e falha no primeiro problema.
  * `ReadLoose` tolera inconsistências e tenta seguir para o próximo registro.
* **Registros deletados**: use `IncludeDeleted: true` para incluir registros marcados como excluídos (`rec["_deleted"] == true`).

## Limitações

* Campos **M (memo)** ainda não são lidos; interpretação de `.DBT/.FPT` em desenvolvimento.
* Biblioteca **somente leitura**: não cria/edita DBF (por enquanto).
* Tipos específicos do Visual FoxPro (ex.: `General`, `Variant`) ainda não são suportados.

## Roadmap

* Suporte completo a campos memo (`.DBT`/`.FPT`).
* APIs de escrita/atualização de registros.
* Suporte ampliado a tipos/versões e validações adicionais.

## Licença

Distribuído sob os termos da [licença MIT](LICENSE).
