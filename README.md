# terraform-registry-builder

## 概要

プライベート Terraform レジストリーとして稼働させるための静的サイトを構築します。

## 実行方法

```
terraform-registry-builder SRC DST
```

* SRC には、新しく構築したプロバイダーバイナリーまたはパッケージがあるディレクトリーを指定します。
* DST には、Terraform レジストリーのネームスペースディレクトリーとして使用するディレクトリーを指定します。

## SRC ディレクトリー内のファイル

SRC ディレクトリー以下に配置するファイル名は以下のフォーマットとしてください。
再帰的に探索を行うので、途中のディレクトリー構造は問いません:

* バイナリーファイルの場合: `terraform-provider-(TYPE)-v(VERSION)_(OS)_(ARCH)`
* zip ファイルの場合: `terraform-provider-(TYPE)-v(VERSION)_(OS)_(ARCH).zip`

バイナリーファイルのみが提供されている場合は、以下の仕様の zip ファイルを作成します:

* 中身はバイナリーファイルだけ。
* 中に含まれるファイルのファイルのモードは 0644 固定
* 中に含まれるファイルのファイルの時刻を 2049年1月1日 0時0分0秒 に固定します。

## DST ディレクトリーへの配置

DST ディレクトリーには、Terraform プロバイダーのネームスペースディレクトリーを指定してください。
例えば配置するプロバイダーが `(DOMAIN)/(NAMESPACE)/(TYPE)` で指定される場合、 `providers/(NANESPACE)` を指定します。

SRC ディレクトリーに `terraform-provider-(TYPE)-v(VERSION)_(OS)_(ARCH)` がある場合、
DST ディレクトリー以下には以下のようなファイルが構築されます:

* `(TYPE)/versions/index.json`
    * 既存のファイルがある場合は既存ファイルに追記する。
* `(TYPE)/(VERSION)/download/(OS)/(ARCH)/index.json`
* `(TYPE)/(VERSION)/download/(OS)/(ARCH)/terraform-provider-(TYPE)-v(VERSION)_(OS)_(ARCH).zip`
* `(TYPE)/(VERSION)/download/(OS)/(ARCH)/terraform-provider-(TYPE)-v(VERSION)_(OS)_(ARCH)_SHA256SUMS`
* `(TYPE)/(VERSION)/download/(OS)/(ARCH)/terraform-provider-(TYPE)-v(VERSION)_(OS)_(ARCH)_SHA256SUMS.sig`

## GPG キーのセットアップ

Terraform レジストリーの仕様上、 GPG による署名が必要になります。
このため利用にあたっては GPG の秘密鍵を用意してください。

### GPG キーの作成方法

新しい GPG キーペアを作成する場合は、以下の手順で実施できます:

```bash
# GPG キーを対話的に生成
gpg --full-generate-key
```

上記コマンド実行後、以下の情報を入力します:
1. キーの種類を選択 (デフォルトの RSA and RSA を推奨)
2. キーサイズを設定 (4096 ビットを推奨)
3. キーの有効期限を設定
4. 名前とメールアドレスを入力
5. コメントを必要に応じて入力
6. パスフレーズを設定

### GPG キー ID の確認

生成されたキーの ID を確認します:

```bash
gpg --list-secret-keys --keyid-format LONG
```

出力例:
```
sec   rsa4096/3AA5C34371567BD2 2023-01-01 [SC] [expires: 2025-01-01]
      1234567890ABCDEF1234567890ABCDEF12345678
uid                 [ultimate] Your Name <your.email@example.com>
ssb   rsa4096/4BB6D45482678BE3 2023-01-01 [E] [expires: 2025-01-01]
```

ここで `3AA5C34371567BD2` がキー ID になります。

### GPG キーのエクスポート

秘密鍵をエクスポートします:

```bash
# ASCII 形式で秘密鍵をエクスポート
gpg --export-secret-keys --armor 3AA5C34371567BD2 > private_key.asc
```

### 環境変数の設定

terraform-registry-builder で利用するために、以下の環境変数を設定します:

```bash
# 秘密鍵ファイルのパスを設定
export TFREGBUILDER_GPG_KEY_FILE=/path/to/private_key.asc

# または、秘密鍵の内容を直接設定することもできます
export TFREGBUILDER_GPG_KEY="$(cat /path/to/private_key.asc)"

# パスフレーズを設定
export TFREGBUILDER_GPG_PASSPHRASE="your_passphrase_here"

# キー ID を設定
export TFREGBUILDER_GPG_ID="3AA5C34371567BD2"
```

### CI/CD 環境での設定

GitHub Actions などの CI/CD 環境では、シークレットとして上記の値を設定し、ワークフロー内で環境変数として利用できます:

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up GPG credentials
        env:
          TFREGBUILDER_GPG_KEY: ${{ secrets.TFREGBUILDER_GPG_KEY }}
          TFREGBUILDER_GPG_PASSPHRASE: ${{ secrets.TFREGBUILDER_GPG_PASSPHRASE }}
          TFREGBUILDER_GPG_ID: ${{ secrets.TFREGBUILDER_GPG_ID }}
        run: |
          # ここで terraform-registry-builder を実行
          ./terraform-registry-builder SRC DST
```

## CI/CD

このプロジェクトでは以下のGitHub Actionsワークフローが設定されています：

### テスト実行

- mainブランチへのプッシュ時
- プルリクエストの作成・更新時
- 自動でGoのテストが実行されます

### リリース

Gitリリースを作成すると以下が実行されます：

1. バイナリビルド
   - 以下のプラットフォーム向けにビルドされます：
     - linux/amd64
     - linux/arm64
     - darwin/arm64
     - windows/amd64
   - ビルドされたバイナリはリリースアセットとして添付されます

2. Dockerイメージ作成
   - マルチプラットフォーム (linux/amd64, linux/arm64) のDockerイメージが作成されます
   - イメージはGitHub Container Registry (ghcr.io)にプッシュされます
