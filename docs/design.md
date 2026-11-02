# Design

## Fluxo de uma execução

```
CLI run / POST /v1/pipelines/{name}/runs / tick do scheduler
  -> service.Execute / service.Trigger
       -> lock do pipeline (ErrAlreadyRunning se já está rodando -> 409 / skip)
       -> grava Run{status: running}
       -> core.Runner.Run(pipeline)
            -> Source.Open(query)                  cursor: colunas já disponíveis
            -> Transformer.Transform(schema)       colunas finais, antes de qualquer dado
            -> Sink.Open(target, colunas)          BEGIN + Strategy.Prepare (DELETE/TRUNCATE/nada)
            -> goroutine leitora: Cursor.Next ---> canal (Prefetch batches) ---> escritora:
                                                    Transform + Writer.Write (INSERTs multi-linha)
            -> sucesso: Writer.Commit   |   qualquer erro / timeout / shutdown: Writer.Rollback
       -> grava Run{status: succeeded|failed, linhas, duração, erro}
       -> libera o lock
```

O alvo nunca observa um estado intermediário: em `replace`, o DELETE e
todos os INSERTs estão na mesma transação; leitores continuam vendo os
dados anteriores (MVCC) até o COMMIT.

## SOLID, onde e por quê

O pedido era "abusar de SOLID". Cada princípio abaixo resolve um
problema concreto do código original, não é decoração.

### S: Single Responsibility

| Tipo / pacote | Única razão para mudar |
|---|---|
| `core.Runner` | como mover batches de um Source para um Sink (streaming, commit/rollback) |
| `sqlsink.Dialect` | sintaxe de um banco (quoting, placeholders, upsert, limites) |
| `sqlsink.Strategy` | semântica de um modo de escrita |
| `service.Service` | regras de execução: lock por pipeline, histórico |
| `config` | formato e validação do arquivo |
| `app` | ligação entre as peças (composition root) |
| `httpapi`, `cli`, `scheduler` | um canal de entrada cada |

No original, `BrandDnbExporter` fazia tudo isso numa função de 100
linhas, repetida 21 vezes: conexão, query, reflection, mapeamento
campo a campo, DELETE, INSERT linha a linha, log.

### O: Open/Closed

Adicionar, sem editar o que existe:

- **um banco de destino SQL**: implementar `sqlsink.Dialect` e registrar
  no `builtin` (ou no seu próprio `main`, via `cli.Env.Connectors`);
- **um destino não-SQL** (S3, Kafka, HTTP): implementar `core.Sink`;
- **um modo de escrita**: implementar `sqlsink.Strategy` e passar
  `sqlsink.WithStrategy(mode, s)`;
- **uma transformação**: `transform.Registry.Register("nome", factory)`;
- **um subcomando**: implementar `cli.Command`.

Nenhuma dessas extensões passa por um `switch` central. O `core`, o
`service` e o `config` não mudam.

### L: Liskov Substitution

Todo `core.Sink` honra o mesmo contrato, verificado pelos testes:
`Rollback` depois de `Commit` é no-op (o Runner sempre faz `defer`),
e nada fica visível antes do `Commit` quando o destino suporta
transação. Quando um destino **não** consegue honrar o contrato, ele
recusa em vez de fingir: o MySQL rejeita `truncate` (TRUNCATE faz
commit implícito e quebraria o rollback), o `jsonl` rejeita `upsert`.
Um substituto que aceitasse e se comportasse diferente seria a violação.

O mesmo vale para `core.Source`: o leitor `sqlsource` funciona com
qualquer driver `database/sql`, e `coretest.Source` é substituível em
todos os testes do Runner, do service e do app.

### I: Interface Segregation

- `core.Writer` (Write/Commit/Rollback) é separado de `core.Sink`
  (Open): quem escreve não abre, quem abre não escreve.
- `core.Pinger` é opcional, detectado por type assertion: um arquivo
  local não precisa fingir um `Ping`.
- `run.Recorder` (só `Save`) é tudo que o `service` precisa;
  `run.Reader` (Get/List) é tudo que a API precisa. `run.Store` junta os
  dois só para quem implementa.
- A API depende de `Catalog`, `Triggerer`, `run.Reader` e `Readiness`,
  quatro interfaces de um ou dois métodos, não do `*service.Service`.
- `sqlsink.Execer` / `Tx` / `TxBeginner` expõem só o subconjunto de
  `*sql.Tx` usado, o que permite testar toda a geração de SQL sem banco.

### D: Dependency Inversion

`core` define as portas e não importa nada além da stdlib. Drivers
(`lib/pq`, `go-sql-driver/mysql`) só são importados em
`connector/builtin`; YAML só em `config`; `net/http` só em `httpapi` e
`app`; go-workers só em `scheduler`. O `service` depende de
`Executor` (implementado por `core.Runner`) e de `run.Recorder`; o
scheduler é a interface `scheduler.Scheduler`, com o go-workers como
um adapter. Trocar o go-workers por um motor de cron, ou o histórico em
memória por uma tabela no Aurora, é trocar uma linha no `app`.

## Concorrência

- **Leitura e escrita sobrepostas**: uma goroutine leitora alimenta um
  canal com capacidade `Prefetch` (padrão 2). Memória máxima:
  `(Prefetch + 1) × batch_size` linhas.
- **Sem goroutine órfã**: em erro de escrita o Runner cancela o contexto
  da leitora, drena o canal e espera o erro final dela antes de fechar o
  cursor. `goleak.VerifyTestMain` falha a suíte se qualquer goroutine
  sobreviver a um teste.
- **Cursor cancelável**: `sqlsource` abre a query com um contexto próprio
  e o cancela no `Close`. Sem isso, abandonar um result set enorme do
  Redshift (lib/pq) significaria ler todas as linhas restantes do socket.
- **Um pipeline por vez**: o lock é por nome de pipeline, no `service`,
  então vale para scheduler, API e CLI ao mesmo tempo. Pipelines
  diferentes rodam em paralelo.

## Shutdown

`serve` recebe SIGTERM:

1. o contexto do scheduler é cancelado: nenhum tick novo;
2. o servidor HTTP para de aceitar requests;
3. runs em andamento (agendados ou via API) continuam, porque rodam com
   um contexto separado (`runCtx`), até `server.shutdown_timeout`;
4. estourado o prazo, `runCtx` é cancelado: as transações fazem rollback.

Passos 2 e 3 correm em paralelo sob o mesmo prazo. Conexões TCP que
nunca enviaram request (health check de load balancer) são fechadas à
força no fim do prazo: o `net/http` só as considera ociosas após 5s e,
sem isso, um shutdown perfeitamente limpo terminava com erro (ver
`TestServe_IdleTCPConnectionDoesNotBlockShutdown`).

## Como adicionar um novo destino

Exemplo: SQL Server.

1. `internal/connector/sqlsink/sqlserver.go`: um tipo que implementa
   `Dialect` (`QuoteIdent` com `[...]`, `Placeholder` com `@pN`,
   `MaxParams` 2100, `TruncateSQL`, `UpsertSuffix`, este último exigindo
   `MERGE`; se não couber num sufixo, implemente uma `Strategy` própria
   para `upsert` e registre com `WithStrategy`).
2. Em `connector/builtin`: importar o driver e
   `r.RegisterSink("sqlserver", sqlSink("sqlserver", sqlsink.SQLServer{}, identity))`.
3. Testes de geração de SQL ao lado de `sink_test.go` e um serviço no
   `docker-compose.yml` + caso em `test/integration`.

Nenhuma outra mudança: config, service, API, scheduler e CLI já
aceitam `driver: sqlserver`.

Para um destino não-SQL (por exemplo, S3 em Parquet), implemente
`core.Sink`/`core.Writer` diretamente (o `jsonl` é o exemplo mínimo,
incluindo commit atômico via rename) e registre com `RegisterSink`.
