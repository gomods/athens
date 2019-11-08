---
title: 使用.netrc文件管理私人仓库
description: 授权Athens访问私有仓库
weight: 10
---

## 通过.netrc文件认证访问私有仓库

1. 创建一个.netrc文件，如下所示：

	`machine <ip or fqdn>`

  	`login <username>`
	
  	`password <user password>`

2. 通过环境变量通知Athens该文件的位置：

	`ATHENS_NETRC_PATH=<location/to/.netrc>`

3. Athens将文件复制到home目录，并覆盖home目录中的任何.netrc文件。或者，如果Athens服务器的主机在home目录中已经存在一个.netrc文件，则身份验证可开箱即用。

## 通过.hgrc认证访问Mercurial私有存储库

1. 创建带有身份验证数据的.hgrc文件

2. 通过环境变量通知Athens该文件的位置

	`ATHENS_HGRC_PATH=<location/to/.hgrc>`

3. Athens将会把文件复制到home目录，并覆盖home目录中的任何.hgrc文件。或者，如果Athens服务器的主机在home目录中已经存在一个.hgrc文件，则身份验证可开箱即用。

