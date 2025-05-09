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
