# Simple migration mysql tool

## Description

This tool is a simple migration tool for mysql databases. It is written in golang and uses the `github.com/go-sql-driver/mysql` package to connect to the database.


## Usage

1. Install

```bash
go build main.go && cp main.exe %GOPATH%\bin\migrator.exe && rm main.exe 
```

2. Command

```bash
migrator install
```


```bash
migrator create {fileName}
```

```bash
migrator run
```
```bash
migration rollback {step}
```
