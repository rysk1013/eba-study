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

- nginxの再起動  
nginx.confの編集時に手早く編集内容を反映できる
    ```
    $ docker exec nginx nginx -s reload
    ```
- watchexec（ホットリロード）を使用しないgoの修正の反映  
go run のプロセスを切るとコンテナが落ちるため、vscodeのターミナルからコンテナごと再起動するのが早いと思われる
    ```
    $ docker restart golang
    ```

## 仕様的なところ

- アカウント情報に関して  
取得したアカウント情報はmainプロセスでメモリ上に保存され有効時間内であれば再取得を行わない  
    - APIから取得するアカウント情報に関してはAPI側で有効時間を24時間に設定して返ってくる  
    - ダミーデータを有効にした場合はローカルのJSON形式のダミーファイルを読み込む（スプレッドシートで作成用の表と式を準備している）  
    ダミーデータから読み込んだアカウント情報の有効時間はプログラム側で6時間で設定
    - ダミーデータを有効にした場合はAPIを呼び出さない
- nginxのキャッシュについて  
nginxから200で返されたレスポンスに関してはnginxで一定時間キャッシュされる(今はとりあえず12時間で設定)

