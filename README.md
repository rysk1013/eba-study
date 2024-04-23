# eba-study

## 環境構築

※Dockerがインストール済みであること

1.リポジトリのクローン

```bash
git clone git@github.com:rysk1013/eba-study.git
```

2.クローンしたディレクトリに移動

```bash
cd eba-study
```

3.コンテナのビルド

```bash
docker-compose -f docker-compose-dev.yml build --no-cache
```

4.コンテナの起動

```bash
docker-compose -f docker-compose-dev.yml up -d
```

## トップページの表示

「todoEBA」から下記URLに遷移する

```
https://localhost:3443
```

## その他

- ローカル環境のコンテナ起動は「docker-compose-dev.yml」を指定する
- ファイルを変更した場合は、Golangのビルドが自動で走るようになっている
