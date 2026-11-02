# Trade-offs e decisões

Esta seção registra problemas reais do código anterior (não
hipotéticos), encontrados lendo o código durante a reescrita, e as
decisões de design que vieram deles. O código anterior foi escrito
rápido, sob pressão, por uma pessoa só; nada aqui é crítica a isso,
é o registro do que uma segunda passagem, com tempo e testes, encontrou.

## 1. Credenciais versionadas no repositório

Senhas do Redshift e do Aurora estavam em constantes Go
(`internal/database/redshifthelper/redshift_connector.go`,
`internal/database/postgreshelper/hotdata_connector.go`), comentadas no
`sqlc.yaml` e no `.env`, que estava commitado (a linha `.env` do
`.gitignore` estava comentada). `AuroraConnector` ainda fazia
`println(connStr)`, imprimindo a senha no stdout a cada conexão.

Agora: nenhum segredo em código ou config; o YAML referencia
`${REDSHIFT_DSN}`, `.env` está no `.gitignore` e só existe um
`.env.example`. **As credenciais antigas continuam no histórico do git
e devem ser rotacionadas**; removê-las do histórico exige reescrevê-lo
(`git filter-repo`), o que não foi feito aqui por ser destrutivo.

## 2. NULL virava zero, `false` ou `0001-01-01`

Os 21 exporters montavam os parâmetros do sqlc assim:

```go
ChurnDate: sql.NullTime{Valid: true, Time: branddnb.ChurnDate.Time},
Isactive:  sql.NullBool{Valid: true, Bool: branddnb.IsActive.Bool},
```

`Valid: true` forçado (148 ocorrências) descarta a informação de NULL
que o Scan acabou de ler: uma marca sem `churn_date` era gravada com
churn em `0001-01-01`, `is_active` NULL virava `false`,
`minimum_balance` NULL virava `"0"`. Corrupção silenciosa, sem erro.

Agora não existe mapeamento campo a campo: o valor lido do driver é o
valor gravado, NULL incluído. O teste de integração
`TestPostgres_ReplaceIsIdempotentAndPreservesValues` verifica uma linha
com seis colunas NULL dos dois lados.

## 3. DELETE + INSERT fora de transação

Cada exporter chamava `DeleteXxxTable()` e depois inseria linha a
linha, cada statement em autocommit. Durante a carga, leitores do Aurora
viam a tabela vazia ou pela metade; se o processo caísse no meio (ou um
INSERT falhasse, ver item 4), a tabela ficava assim até a próxima
execução.

Agora o modo `replace` faz DELETE e INSERTs numa única transação, com
rollback em qualquer erro. `TestPostgres_FailedRunLeavesPreviousDataIntact`
força uma violação de CHECK na linha 2.400 de 2.500 e confere que os
dados anteriores continuam lá, intactos.

`DELETE` em vez de `TRUNCATE` como padrão é deliberado: no PostgreSQL
`TRUNCATE` é transacional mas pega `ACCESS EXCLUSIVE`, bloqueando
leitores até o commit; `DELETE` não bloqueia (custo: mais bloat, que o
autovacuum resolve). `truncate` existe como modo explícito para quem
prefere velocidade a disponibilidade de leitura.

## 4. Erros ignorados e `log.Fatal` dentro do loop

```go
rows, err := db.QueryContext(ctx, REDSHIFT_QUERY)
if err != nil { util.Sugar.Errorf(...) }   // segue com rows == nil
...
err := rows.Scan(columns...)
if err != nil { log.Fatal(err) }           // derruba o processo inteiro
```

Uma query com erro seguia adiante e causava panic de nil pointer; um
Scan com erro matava todos os workers do processo, no meio de uma carga
não transacional (item 3). Falhas de INSERT eram logadas e o loop
continuava, então uma execução "bem-sucedida" podia ter perdido linhas.

Agora todo erro interrompe a execução, faz rollback, e chega ao
chamador: status `failed` no histórico, `409/500` na API, exit code 1
no CLI, retry no scheduler.

## 5. Scan por reflection dependente da ordem das colunas

`select *` era escaneado nos campos da struct por posição
(`reflect.ValueOf(&entity).Elem().Field(i)`). Uma coluna adicionada ou
reordenada no Redshift deslocava todos os valores silenciosamente (ou
quebrava no Scan, item 4).

Agora as linhas são `[]any` com os nomes de coluna reportados pelo
driver, e o INSERT lista as colunas por nome. Divergências de nome
viram erro do banco, não dados trocados.

## 6. Rota HTTP chamando o exporter errado

```go
vmBrandsController := controllers.NewVMBrandWalletHistoryController()
router.POST("/vm_brands/", vmBrandsController.ImportAllRows)
```

`POST /vm_brands/` disparava a exportação de `brand_wallet_history`.
Copy-paste entre 21 blocos quase idênticos é exatamente onde esse erro
nasce; com um único endpoint parametrizado
(`POST /v1/pipelines/{name}/runs`) ele deixa de ser possível.

## 7. Execuções concorrentes do mesmo exporter

Cada `POST` disparava `go models.XxxExporter(ctx)` sem nenhuma guarda.
Dois POSTs seguidos rodavam dois "DELETE + INSERT" intercalados na
mesma tabela: linhas duplicadas, ou uma tabela esvaziada pelo DELETE do
segundo no meio da carga do primeiro.

Agora o `service` mantém um lock por pipeline, compartilhado por API,
scheduler e CLI: a segunda requisição recebe `409 Conflict`, o tick do
scheduler é pulado (sem retry).

## 8. Inserção linha a linha

Um round trip por linha. Agora INSERTs multi-linha, com o número de
linhas por statement limitado por `rows_per_statement` e pelo limite de
bind parameters do dialeto (65.535 em PostgreSQL e MySQL). `COPY` no
PostgreSQL seria ainda mais rápido e está no roadmap; não entrou agora
porque não suporta `upsert` e exigiria um segundo caminho de escrita.

## 9. Dependências

De 64 módulos no go.mod (gin, viper, fsnotify, swaggo, zap, uber/config,
confluent-kafka-go) para 6 diretos: `lib/pq`, `go-sql-driver/mysql`,
`yaml.v3`, `go-workers`, `google/uuid`, `goleak` (só testes).

- gin → `net/http` da stdlib (roteamento por método e path params desde
  o Go 1.22).
- zap → `log/slog`.
- viper + fsnotify + uber/config → o pacote `config` (~400 linhas, metade comentários e tipos), com
  validação estrita. O hot reload do viper relia o `.env`, mas as
  conexões eram lidas uma vez em variáveis de pacote, então o reload
  nunca tinha efeito.
- confluent-kafka-go não era usado por nenhum caminho ativo e exigia
  cgo (librdkafka); removê-lo é o que permite a imagem `FROM scratch`.

## 10. go-workers como adapter, não como dependência do domínio

O scheduler original passava `*sql.Rows` entre steps pelo
`TaskParams` do go-workers, acoplando o pipeline ao formato interno da
biblioteca. Agora o go-workers é usado só para o que faz bem (agenda,
retries, recover de panic) atrás da interface `scheduler.Scheduler`.

Duas consequências conhecidas: o go-workers loga com seu próprio
logger JSON em stdout (`worker finished`), separado do `slog` do
aurora-exporter em stderr; e o intervalo tem resolução de segundos, sem
expressões cron. Ambos se resolvem trocando o adapter, sem tocar no
resto.

## 11. `sanitize_utf8` como transformação, não como regra fixa

O código antigo tinha `FilterInvalidUTF8` em `pkg/util`, usada em
alguns exporters e não em outros. O PostgreSQL rejeita NUL (`0x00`) em
`text` e o Redshift os devolve, então uma linha ruim derrubava a carga
inteira. Virou uma transformação nomeada, ligada por padrão nos
exemplos via `defaults.transforms`, e desligável por pipeline.
