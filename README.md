# kiso-lab-support-tool

## 使い方

### dockerの場合

#### DockerHubからimageを持ってくる場合

DockerHubから持ってくる

```shell
docker pull shakku/kiso-lab-support-tool
```

持ってきたimageでコンテナ起動

```shell
docker run -e USER_ENV="user" -e PASSWORD_ENV="password" -p 8080:8080 -d kiso-lab-support-tool
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

作成したimageでコンテナ起動

```shell
docker run -e USER_ENV="user" -e PASSWORD_ENV="password" -p 8080:8080 -d kiso-lab-support-tool
```

### goコマンドで実行する場合

pythonの実行パスを以下にする。
pypdfというパッケージを使用しているため、環境に合わせてpythonのバージョン指定を変更する。

```go:main.go
PythonPath = "python3"
```

```shell
go run main.go -user ユーザー名 -password パスワード
```

## メモ

- 汎用的CSSを使って、見た目よく ok
- Dockerfileで自動でサーバー起動するように ok
- 自動サーバー起動時にパスワードのフラグどうするか ok
- merge.pdfが更新されたら、/pdfのページを自動更新したい ok
- 質問機能
- 運営からのメッセージ機能 ok
- 実行ポートをフラグで指定
