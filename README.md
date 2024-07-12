# kiso-lab-support-tool

## 使い方？

```shell
docker image build -t kiso-lab-support-tool .
```

```shell
docker run -e USER_ENV="user" -e PASSWORD_ENV="password" -p 8080:8080 -d kiso-lab-support-tool
```

- 汎用的CSSを使って、見た目よく ok
- Dockerfileで自動でサーバー起動するように ok
- 自動サーバー起動時にパスワードのフラグどうするか ok
- merge.pdfが更新されたら、/pdfのページを自動更新したい
- 質問機能
- 運営からのメッセージ機能 ok
