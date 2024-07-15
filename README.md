# kiso-lab-support-tool

## 起動方法

### dockerの場合

#### DockerHubからimageを持ってくる場合

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
pypdfというパッケージを使用しているため、環境に合わせてpythonのバージョン指定を変更する。

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
- '/pdf'ページにアクセスすると、maxPageまでのページのpdfが表示され、maxPageが更新されると、pdfが自動で更新され、新たに指定されたmaxPageまでのページを見れる。
- Managementからタイトルと本文を入力して全体にメッセージを送信することができる。
- '/message'にアクセスすると、管理者から送信されたメッセージが一覧表示され、新たにメッセージが送信されると自動で更新されて表示される。

## メモ

- 汎用的CSSを使って、見た目よく ok
- Dockerfileで自動でサーバー起動するように ok
- 自動サーバー起動時にパスワードのフラグどうするか ok
- merge.pdfが更新されたら、/pdfのページを自動更新したい ok
- 質問機能
- 運営からのメッセージ機能 ok
- 実行ポートをフラグで指定 ok
