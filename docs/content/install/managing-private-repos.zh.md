---
title: 使用 .netrc 文件管理私有仓库
description: 授权 Athens 访问私有仓库
weight: 10
---

## 通过 .netrc 文件认证访问私有仓库

1. 创建一个 .netrc 文件，如下所示：

	`machine <ip or fqdn>`

  	`login <username>`

  	`password <user password>`

2. 通过环境变量告诉 Athens 该文件的位置：

	`ATHENS_NETRC_PATH=<location/to/.netrc>`

3. Athens 会将文件复制到 home 目录，并覆盖 home 目录中的任何 .netrc 文件。或者，如果 Athens 服务器的主机在 home 目录中已经存在一个 .netrc 文件，则身份验证开箱即用。

## 通过 .hgrc 认证访问 Mercurial 私有仓库

1. 创建带有身份验证数据的 .hgrc 文件

2. 通过环境变量告诉 Athens 该文件的位置

	`ATHENS_HGRC_PATH=<location/to/.hgrc>`

3. Athens 会将文件复制到 home 目录，并覆盖 home 目录中的任何 .hgrc 文件。或者，如果 Athens 服务器的主机在 home 目录中已经存在一个 .hgrc 文件，则身份验证开箱即用。
