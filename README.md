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
