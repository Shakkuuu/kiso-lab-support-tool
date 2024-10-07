# kiso-lab-support-tool

## 概要

[https://qiita.com/Shakku/items/9969380a4313bc33ce95](https://qiita.com/Shakku/items/9969380a4313bc33ce95)

イベントや講演会、ハンズオンを行う際に便利なアプリケーション。

スライドや PDF の資料を使用する際に、参加者の手元で資料を見ることができ、前ページを見返すことや、スライドの進行状況に応じて先のページのネタバレやカンニングが起きないように資料を共有することができる。
資料を見ながらの手入力が大変な文字列（コマンドやプログラムなど）を参加者に入力させたいとき、メッセージ機能で管理者から送信することで、参加者はコピー&ペーストをすることができる。また、管理者からの連絡や補足、資料訂正などさまざまな用途で使用できる。
管理者が資料の進行状況を更新したりメッセージを送信すると、それぞれ自動的にリアルタイムでクライアント側にも反映される。

## 起動方法

### dockerの場合

#### DockerHubからimageを持ってくる場合

[https://hub.docker.com/r/shakku/kiso-lab-support-tool](https://hub.docker.com/r/shakku/kiso-lab-support-tool)

DockerHubから持ってくる

```shell
docker pull shakku/kiso-lab-support-tool
```

持ってきたimageでコンテナ起動。
環境変数指定でBasicAuthのユーザとパスワードをstringで指定し、サーバ起動ポートをintで指定する。

```shell
docker run -e USER_ENV="user" -e PASSWORD_ENV="password" -e PORT_ENV=8080 -p 8080:8080 -d shakku/kiso-lab-support-tool
```

#### buildからする場合

pythonの実行パスを以下にする

```go:main.go
PythonPath = "/opt/venv/bin/python3"
```

build

```shell
docker image build -t kiso-lab-support-tool .
```

作成したimageでコンテナ起動。
環境変数指定でBasicAuthのユーザとパスワードをstringで指定し、サーバ起動ポートをintで指定する。

```shell
docker run -e USER_ENV="user" -e PASSWORD_ENV="password" -e PORT_ENV=8080 -p 8080:8080 -d kiso-lab-support-tool
```

### goコマンドで実行する場合

pythonの実行パスを以下にする。
PyMuPDFやPillowというパッケージを使用しているため、環境に合わせてpythonのバージョン指定を変更する。

```go:main.go
PythonPath = "python3"
```

引数でBasicAuthのユーザとパスワードをstringで指定し、サーバ起動ポートをintで指定する。

```shell
go run main.go -user user -password password -port 8080
```

## 使い方

- '/management'にアクセスし、発表資料のpdfをアップロードする。
- 発表の進捗に合わせて、maxPageの数字を変更する。
- '/document'ページにアクセスすると、maxPageまでのページの資料が表示され、maxPageが更新されると、資料が自動で更新され、新たに指定されたmaxPageまでのページを見れる。
- Managementからタイトルと本文を入力して全体にメッセージを送信することができる。
- '/message'にアクセスすると、管理者から送信されたメッセージが一覧表示され、新たにメッセージが送信されると自動で更新されて表示される。
- ManagementからManagementMessageにアクセスし、メッセージを削除できる。（一般ユーザーのメッセージ一覧ではDeleteボタンは表示されない。）

## 注意

- pdfのファイルサイズは100MBまで
- ページ数は10000ページまで
- メッセージのタイトルは50文字まで
- メッセージの本文は10000文字まで

## メモ

- 質問機能
- ファイル配布
- ページ送りのたびにSSEクライアントに参加しているため、たまに読み込みがものすごく遅くなる
- 汎用的CSSを使って、見た目よく ok
- Dockerfileで自動でサーバー起動するように ok
- 自動サーバー起動時にパスワードのフラグどうするか ok
- merge.pdfが更新されたら、/pdfのページを自動更新したい ok
- 運営からのメッセージ機能 ok
- 実行ポートをフラグで指定 ok
- 見た目改良 ok
- messageをdb管理に変更 ok
- ファイル分け ok
- messageの並び順を最新順に ok
- echo v4にする ok
- バリデーション ok
- golang 1.22.5 ok
- ファイルサイズ制限？ ok
- サーバ側でpdfしかダメにする ok
- ディレクトリの権限 ok
- ファイル形式の確認 ok
- XSS対策 ok
- MessageのContentをHTML表示できるようにする ok
- 範囲指定して、ここからここまでmergeとかにする？ no
- 左側にプレビューで表示可能ページまで小さくプレビューとそのページのリンクを表示して、最初の時みたいに次へ戻るボタンでページ送りするようにする？最大ページまでいったら次へが表示されないようにしたい ok
- if文でnextpageとかbackpageのaを表示するかどうか。 ok
- 範囲外のページ指定した時にview-pdf内をみてその数字のpdfがなければ404にする。 ok 1ページに移行するようにした
- pdfファイルのサイズ圧縮とかして小さくする no Web上の変換サイトばっかりで、CLIでGhostscriptというものがあったが、脆弱性があるらしい
- 最大ページ更新のたびに再読み込みだから、ページ数が多いと読み込み遅くなる ok
- pdfじゃなくで、jpegにして、imgで表示させる？imgだと読み込みの優先順位とかができるらしい ok jpgの方が軽かった
- 結果的に、リロードのたびに1ページ目に戻されてた問題が解決した。
- プレビューをハンバーガーにして隠せるようにする。 ok
- SSEのLock ok
- DockerfileのCMDの書き方(# CMD ["./main", "-user", "$USER_ENV", "-password", "$PASSWORD_ENV", "-port", "$PORT_ENV"])でうまく起動できないのを確認する。 ok
- リクエスト数は減らせてないが、ログを分けてとりあえず対応
- dockerfileサイズ削減する ok
- 開発者モードからDelete見えない？メッセージのIDとか ok大丈夫
- SSEをどっちもupdateにしてるから、どっちもリロードされちゃう？ 修正ok
- ハンバーガー内の画像はハンバーガー開かれるまで読み込まないようにしたい ok
- gzip化してリクエスト数を減らす no

docker image build --platform linux/amd64 -t kiso-lab-support-tool .
