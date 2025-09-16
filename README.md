# dbfmini

Leitor **puro Go** para arquivos **dBASE/DBF** (sem dependências de bibliotecas DBF de terceiros).  
Suporta campos: **C, N, F, Y, L, D, I, T, B**.  
Campos **M (memo)**: _TODO_ (esqueleto preparado para `.DBT/.FPT`).

> Codificações de texto: **CP850, CP437, CP1252, ISO-8859-1, UTF-8** via `golang.org/x/text`.

## Instalação

```bash
go get github.com/<seu-usuario>/dbfmini@v0.1.0
