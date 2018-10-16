---
title: Configuring Authentication
description: Configuring Authentication on Athens
weight: 1
---

## Authentication

## SVN private repositories

1. Subversion creates an authentication structure in 
        
        ~/.subversion/auth/svn.simple/<hash>

2. In order to properly create the authentication file for your SVN servers you will need to authenticate to them and let svn build out the proper hashed files.
	
		$ svn list http://<domain:port>/svn/<somerepo>
		Authentication realm: <http://<domain> Subversion Repository
		Username: test
		Password for 'test':

3. Once we've properly authenticated we want to share the .subversion directory with the proxy server in order to reuse those credentials.  Below we're setting it as a volume on our proxy container.

**Bash**

```bash
export ATHENS_STORAGE=~/athens-storage
export ATHENS_SVN=~/.subversion
mkdir -p $ATHENS_STORAGE
docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
   -v $ATHENS_SVN:/root/.subversion \
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens \
   -e ATHENS_STORAGE_TYPE=disk \
   --name athens-proxy \
   --restart always \
   -p 3000:3000 \
   gomods/proxy:latest
```

**PowerShell**

```PowerShell
$env:ATHENS_STORAGE = "$(Join-Path $pwd athens-storage)"
$env:ATHENS_SVN = "$(Join-Path $pwd .subversion)"
md -Path $env:ATHENS_STORAGE
docker run -d -v "$($env:ATHENS-STORAGE):/var/lib/athens" `
   -v "$($env:ATHENS-SVN):/root/.subversion" `
   -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
   -e ATHENS_STORAGE_TYPE=disk `
   --name athens-proxy `
   --restart always `
   -p 3000:3000 `
   gomods/proxy:latest
```