# Go 読解ノート

## Go の基礎(読むための最小限。読解のたびに追記する)

| 書き方 | 意味 | 出てくる場所 |
|---|---|---|
| `package product` / `import "…/internal/product"` | 1フォルダの中の .go ファイルは全部同じパッケージに属する(フォルダ内で混在できない)。同じパッケージのファイル同士は import なしで互いの名前を使える(repository.go が product.go の `ValidationError` をそのまま使えるのはこのため)。import できる数には制限がなく、他のパッケージは import して `product.Xxx` で使う | 各ファイルの先頭 |
| 名前の先頭が大文字(`FindByID`、`ErrNotFound`) | パッケージの外から使える(public)。小文字(`db`、`writeJSON`)は中だけ(private) | 全体 |
| `internal/` | その親フォルダの中からしか import できない(Go のルール) | `internal/` |
| `type Repository struct { db *sql.DB }` | 構造体(クラスのデータ部分)。Go にクラスや継承はない | repository.go |
| `func (r *Repository) FindByID(…)` | `Repository` のメソッド。`r` は this のようなもの | repository.go |
| `*T` / `&x` | `*T` は「T の場所(ポインタ)」の型、`&x` は x の場所を取る。コピーせず同じものを共有したいときに使う | `&p.ID`、`&ValidationError{…}` |
| `x := …` | 型を書かずに変数を作る(型は右辺から決まる)。`var p Product` は全項目がゼロ値の変数 | 全体 |
| `(Product, error)` | 戻り値を2つ返せる。慣例として最後が error | repository.go |
| `if err != nil { return … }` | **Go には例外(try/catch)がない**。失敗は error という普通の値で返し、呼んだ側がその場で確認する。nil は「エラーなし」 | 全体 |
| `errors.New("…")` | エラーの値を1つ作る。`ErrNotFound` のように変数に入れて「目印」として使う(sentinel error) | product.go |
| `Error() string` を持つ型 | それだけで error として扱える(インターフェース。②で詳しく)。`ValidationError` は項目名などの情報を持てる error | product.go |
| `fmt.Errorf("…: %w", err)` | **元のエラーを中に包んで**、説明を足した新しいエラーを作る。`%v` は文字列として埋め込むだけで、中身は包まない | repository.go 29行目 |
| `errors.Is(err, ErrNotFound)` | 包みを外しながら、中に `ErrNotFound` が入っているかを調べる | handler.go |
| `errors.As(err, &verr)` | 包みを外しながら、`*ValidationError` 型のものを探し、見つかれば `verr` に取り出す | handler.go |
| `type Store interface { FindByID(…) (Product, error) }` | インターフェース = 「このメソッドを持っているもの」という条件だけを書いた型。中身(処理)は書かない | service.go |
| (implements を書かない) | **Go では「満たします」と宣言しない**。メソッドが揃っていれば自動でそのインターフェースとして扱える(`*Repository` も `*fakeStore` も `FindByID` を持つので `Store` になれる)。TypeScript の型の付き方に近い | service.go / service_test.go |
| `context.Context` / `ctx` | リクエストに付いて回る「締め切りと中止の合図」。関数の第1引数で下へ下へと渡していくのが決まり | 全体 |
| `ctx, cancel := context.WithTimeout(ctx, 2*time.Second)` | 元の ctx に「2秒後に中止」を足した新しい ctx を作る。`cancel` は早めに後片付けする関数 | service.go |
| `defer cancel()` | `defer` = 「この関数を抜けるときに必ず実行する」予約(JS の finally に近い)。`main.go` の `defer db.Close()` も同じ | service.go、main.go |
| `<-ctx.Done()` / `ctx.Err()` | 締め切りが来ると `Done()` に合図が届く。そのとき `ctx.Err()` は `context.DeadlineExceeded` | service_test.go |
| `select { case …: case …: }` | 複数の合図のうち、先に届いたほうを1つ実行する(テストで「2秒待つ」と「締め切り」の早い者勝ちに使用) | service_test.go |
| `xxx_test.go` / `func TestXxx(t *testing.T)` | テストのファイルと関数の決まった名前。`go test ./...` で全部実行される | service_test.go |
| ゴルーチン(`go f()`) | 軽量な並行処理。`http.ListenAndServe` は**リクエスト1件ごとにゴルーチンを1つ作って** handler を呼ぶ(コードに `go` は出てこないが、標準ライブラリの中でやっている) | main.go(見えない所) |

## 読解①(2026-10-07): 最小の API(/health と商品1件)

### ファイルの役割とリクエストの流れ

| ファイル | 役割 | たとえると |
|---|---|---|
| `compose.yaml` | PostgreSQL 16 を Docker で起動する設定(ポート 5433) | — |
| `db/init.sql` | DB の初回起動時に1回だけ実行。`products` テーブルを作り2件入れる | — |
| `go.mod` / `go.sum` | このプロジェクトの名前と、使う外部ライブラリ(pgx = PostgreSQL のドライバ)の一覧 | package.json / package-lock.json |
| `cmd/api/main.go` | 起動の入口。DB 接続を用意し、部品を組み立て、8080番で待ち受ける | 店を開ける店長 |
| `internal/httpapi/handler.go` | HTTP の窓口。URL を振り分け、入力を確認し、結果をステータスコードと JSON にして返す | 受付 |
| `internal/product/product.go` | 商品の形(`Product`)と、起こりうる失敗(`ErrNotFound`、`ValidationError`)の定義 | 商品カードと失敗の種類の一覧 |
| `internal/product/repository.go` | DB に SQL を投げて商品を1件取る | 倉庫係 |

補足(質問への回答):
- `go.mod` は「使うライブラリと版」の宣言(人が読む)、`go.sum` は各版の中身のハッシュ値(改ざん・取り違え検知用。機械が使う)。どちらも Git に入れる。`// indirect` は直接 import していないが pgx が使っているもの
- `compose.yaml` の `volumes: - ./db/init.sql:/docker-entrypoint-initdb.d/init.sql:ro` は、手元のファイルをコンテナの中の場所に見せる設定(`手元:コンテナ内:ro`、ro = 読み取り専用)。postgres イメージは、DB が空の初回起動時だけ `/docker-entrypoint-initdb.d/` の SQL を実行する。init.sql を直しても2回目以降の起動では反映されない(`docker compose down -v` で DB を消してから起動し直す)
  - 左(手元のパス)は自分で自由に決める。右(コンテナ内のパス)は**イメージの作者が決めた場所**で、こちらは合わせるだけ。postgres イメージの起動スクリプト(`/usr/local/bin/docker-entrypoint.sh`)に「DB が空なら `/docker-entrypoint-initdb.d/*` を実行する」と書かれている(10/07 にコンテナ内で確認)
  - 認識の修正(10/07): init.sql は「表示」ではなく「実行」される。フォルダも PostgreSQL も、こちらが作ったのではなく**イメージに最初から入っている**(イメージ = PostgreSQL をインストール済みのディスクの写し、コンテナ = それを起動したもの)。こちらは init.sql を見せただけ
  - `volumes` には2種類ある。**bind mount**(手元のファイル・フォルダを見せる。今回の init.sql)と **volume**(Docker が管理するデータの保存場所)。DB の記憶領域は後者で、コンテナ内の `/var/lib/postgresql/data` に付いている。compose.yaml に書いていないのに付いているのは、postgres イメージが「ここは volume にする」と宣言しているため(名前のない匿名 volume が自動で作られる。`docker inspect` で確認)。コンテナを消しても volume は残るので DB の中身も残り、`down -v` で volume ごと消すと次の起動が「初回」に戻る
  - レビュー観点: 本番用の compose では DB の volume に名前をつけて明示する(`dbdata:/var/lib/postgresql/data`)。匿名 volume は、どれが何の volume か分からなくなる(`docker volume ls` に意味のない ID が並ぶ)
  - postgres 公式イメージを使うなら、どのプロジェクトでも同じ場所。mysql・mariadb の公式イメージも同じ名前の慣習を採用している。ただし Docker 全体の決まりではなく、イメージごとの決まり(redis などにはない)。使うイメージの Docker Hub のページ(「Initialization scripts」の節)で確認する
- DB の形(表と列)を決めているのは `init.sql`。`product.go` はアプリの中での商品の形、`repository.go` は型ではなく「DB の行 ⇔ `Product`」を変換して運ぶ処理

`GET /products/1` の流れ:

1. `main.go` の `ListenAndServe` がリクエストを受け、`Routes()` の振り分け表に渡す
2. `"GET /products/{id}"` に一致 → `getProduct` が呼ばれる
3. `getProduct` が URL から `"1"` を取り出し、数値 `1` に変換(できなければ 400)
4. `repository.FindByID(ctx, 1)` を呼ぶ
5. `FindByID` が `SELECT … WHERE id = $1` を実行し、結果の列を `Product` に詰めて返す(0件なら `ErrNotFound`)
6. `getProduct` がエラーの種類で分岐(400 / 404 / 500)、成功なら `Product` を JSON にして 200 で返す

### パッケージ構成(main、internal/、import の向き)

コード: `playground/go-reading-01/`(`main → httpapi → product` の一方向)

- Q: `product` が `httpapi` を import して、repository の中で 404 を返したら何が困るか
- 最初の答え: 共通のエラー処理が書けなくなり、同じ定義をそれぞれで書く必要が出る → 半分。核心ではなかった
- 整理:
  1. **そもそもコンパイルできない**。`httpapi` はすでに `product` を import しているので、逆向きも足すと循環 import になる(Go は禁止)。Go では「依存は一方向」が言語のルールで強制される
  2. **HTTP 以外から使えなくなる**。notifier ワーカー(SQS)やバッチは「404」という概念を持たない。repository は「見つからない(`ErrNotFound`)」という事実だけを返し、それを 404 にするか、リトライするか、ログだけにするかは呼ぶ側が決める
  3. 結果として、`ErrNotFound` の定義は `product` に1つだけで済み、HTTP の対応表(`ErrNotFound` → 404)も `httpapi` に1つだけで済む
- 1文で: 「下の層は事実だけを返し、それをどう見せるかは上の層が決める。だから import は上から下への一方向にする」


### エラー処理(if err != nil、%w、errors.Is / errors.As)

- Q: `ErrNotFound` を `%v` で包んで返したら、`/products/999` は何番になるか。`%w` なら?
- 最初の答え: writeJSON に入る / %w は分からない → 不正解(どの case も最後は writeJSON を呼ぶ。問いは「どの case か」)
- 正解: `%v` → `errors.Is` も `errors.As` も false → `default` → **500**(商品がないだけなのに「サーバーの故障」になり、ログにもエラーが出る)。`%w` → `errors.Is` が true → **404**
- 実験(10/07): 両方とも `Error()` の文字列は `find product 999: product not found` で同じ。違いは中身で、`errors.Unwrap` すると `%v` は `<nil>`(元のエラーは捨てられている)、`%w` は `product not found`(中に入っている)
- Q: 500 のとき、なぜ詳しい内容を利用者に返さず "internal error" だけにするのか
- 最初の答え: サーバー側の問題なのでクライアントは対処できない。一律 500 を返す → 半分(理由の1つ目は正しい)
- 足りなかった点: **内部情報が漏れる(セキュリティ)**。DB を止めたときのエラーには、DB のユーザー名(`app`)、DB 名(`shop`)、接続先のホストとポート(`127.0.0.1:5433`)、使っているドライバ(pgx)が入っていた。攻撃者にとっては、どこを狙えばよいかの手がかりになる(SQL のエラーなら表名・列名まで出ることがある)
- 1文で: 「500 の詳細は、利用者には対処できず、攻撃者には手がかりになるので、外にはひとこと(+問い合わせ用の ID)だけ返し、詳細はログに残す」
- JS で言うと: `%v` = `new Error("find product: " + err.message)`(文字だけ)、`%w` = `new Error("find product: …", { cause: err })`(本体を cause として持つ)。`errors.Is` は文字列ではなく、cause をたどって本体と `===` で一致するかを見る。確認問題(`%w` なら `errors.Is` は?)→ true と正解(10/07)
- 教訓: **見た目(ログの文字列)が同じなので、`%v` の間違いはログを見ても気づけない**。AI のコードをレビューするときは、エラーを包む `fmt.Errorf` が `%w` になっているかを確認する(review-checklists/app-go.md に入れる)



## 読解②(2026-10-08): インターフェース、context、ゴルーチン

コード: `playground/go-reading-01/` に `internal/product/service.go` と `service_test.go` を追加。handler は Repository ではなく Service を呼ぶように変更

### 追加・変更したファイル

| ファイル | 役割 | たとえると |
|---|---|---|
| `internal/product/service.go`(新規) | 業務の層。Store(インターフェース)を受け取り、締め切り(2秒)を付けて呼ぶ | 受付と倉庫の間に立つ担当者。「2秒で返事がなければ諦める」と決める |
| `internal/product/service_test.go`(新規) | DB を使わない偽物の Store(`fakeStore`)を渡して Service を試す | 練習用の倉庫係 |
| `handler.go`(変更) | `*product.Repository` → `*product.Service`、`FindByID` → `Get` | — |
| `main.go`(変更) | Repository を作り → それを Service に渡し → Service を Handler に渡す(組み立て) | — |

`GET /products/1` の流れ(①から変わった所は太字):

1. `ListenAndServe` がリクエストを受け、**このリクエスト専用のゴルーチンで** `getProduct` を呼ぶ
2. `getProduct` が id を数値にする(できなければ 400)
3. **`service.Get(r.Context(), 1)` を呼ぶ。`r.Context()` は「利用者が接続を切ったら中止」の合図付き**
4. **`Get` がそれに「2秒で中止」を足した ctx を作り、`store.FindByID(ctx, 1)` を呼ぶ。`store` の中身は本番では `*Repository`、テストでは `*fakeStore`**
5. `FindByID` が ctx 付きで SQL を実行。**2秒を過ぎたらドライバが問い合わせを打ち切り、エラーが返る**
6. `getProduct` がエラーの種類で 400 / 404 / 500 に分ける

確認したこと(10/08): `go test -v ./...` で2本とも成功。`TestGet_TimesOut` は、2秒かかる偽物に 0.1秒の締め切りを付けると、約0.1秒で `context.DeadlineExceeded` が返った

### 穴埋め: 締め切り切れを 504 で返す(10/08)

- 予測: 穴を埋める前のテストは何番が返るか → 実行して 500。理由は説明を受けて理解した: 締め切り切れ(`context.DeadlineExceeded`)を確かめる case がなく、`default` に落ちるから
- 自分で書いた(handler.go): `case errors.Is(err, context.DeadlineExceeded):` → `writeJSON(w, http.StatusGatewayTimeout, …)`、import に `"context"`。`go test` 成功
- 意味: 「遅すぎて諦めた」はサーバーの故障(500)ではなく時間切れ(504)。利用者・監視の側で「再試行すればよいか」の判断が変わる

### 壊す: 偽物の倉庫係が %v で包んだら(10/08)

- 予測: 分からなかった → 説明を受けてから実行
- 結果: `fmt.Errorf("find product: %v", ctx.Err())` にすると **500**、`%w` にすると **504**(テスト成功)
- 理由: `%v` は元のエラーを文字に書き写して本体を捨てる。`errors.Is` は文字ではなく本体を探すので見つからず、`default` に落ちる
- 教訓: handler で正しく case を書いても、**下の層が1か所でも `%v` で包むと素通りされる**。しかもログの文字は同じなので気づけない。レビューでは「エラーを包む所はすべて `%w` か」を確かめる

## なぜ API を Go にするのか(2026-10-08)

1. **読みやすさ**: 言語の機能が少なく誰が書いても似たコードになるので、AI が書いたコードをレビューしやすい
2. **性能**: リクエストごとに軽いゴルーチンで並行処理し、少ないメモリで多くの同時アクセスを捌ける
3. **運用**: ビルドすると1つの実行ファイルになり、コンテナが小さく起動も速いので、ECS で扱いやすい

(RAG は Python: 埋め込み・検索・評価などの AI 用ライブラリが Python に集まっているため、役割で言語を分ける)

補足(説明を受けた内容):
- 「短くて書きやすい」とは言わない。Go は `if err != nil` などでむしろ長くなりがちで、突っ込まれる。強みは「短い」ではなく「機能が少なく書き方がそろう」こと
- 性能: OS のスレッドは1本で数 MB、ゴルーチンは数 KB から始まる。Go は事前に機械語へコンパイルするので、1行ずつ解釈する言語より速い
- 運用: Python・Node は本番に本体とライブラリ一式が要るが、Go は実行ファイル1つで動く。イメージが小さく(数十 MB 程度)、起動が速く、本番だけ版が違う事故が起きにくい
