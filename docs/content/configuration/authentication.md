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

3. Once we've properly authenticated we want to share the .subversion directory with the Athens proxy server in order to reuse those credentials.  Below we're setting it as a volume on our proxy container.

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

## Bazaar(bzr) private repositories

1. Bazaaar config files are located in

  - Unix
       
        ~/.bazaar/
  - Windows
        
        C:\Documents and Settings\<username>\Application Data\Bazaar\2.0

  - You can check your location using
  
        bzr version

2. There are 3 typical configuration files
   
  - bazaar.conf
    - default config options
  - locations.conf
    - branch specific overrides and/or settings
  - authentication.conf
    - credential information for remote servers

3. Configuration file syntax

  - \# this is a comment
  - [header] this denotes a section header
  - section options reside in a header section and contain an option name an equals sign and a value
    - EXAMPLE:
        
          [DEFAULT]
          email = John Doe <jdoe@isp.com>

4. Authentication Configuration
   
     Allows one to specify credentials for remote servers.
     This can be used for all the supported transports and any part of bzr that requires authentication(smtp for example).
     The syntax obeys the same rules as the others except for the option policies which don't apply.

     Example:

       [myprojects]
       scheme=ftp
       host=host.com
       user=joe
       password=secret

       # Pet projects on hobby.net
       [hobby]
       host=r.hobby.net
       user=jim
       password=obvious1234

       # Home server
       [home]
       scheme=https
       host=home.net
       user=joe
       password=lessobV10us

       [DEFAULT]
       # Our local user is barbaz, on all remote sites we're known as foobar
       user=foobar

       NOTE: when using sftp the scheme is ssh and a password isn't supported you should use PPK

       [reference code]
       scheme=https
       host=dev.company.com
       path=/dev
       user=user1
       password=pass1

       # development branches on dev server
       [dev]
       scheme=ssh # bzr+ssh and sftp are availablehere
       host=dev.company.com
       path=/dev/integration
       user=user2

       #proxy
       [proxy]
       scheme=http
       host=proxy.company.com
       port=3128
       user=proxyuser1
       password=proxypass1

5. Once we've properly setup our authentication we want to share the bazaar configuration directory with the Athens proxy server in order to reuse those credentials.  Below we're setting it as a volume on our proxy container.

    **Bash**

    ```bash
    export ATHENS_STORAGE=~/athens-storage
    export ATHENS_BZR=~/.bazaar
    mkdir -p $ATHENS_STORAGE
    docker run -d -v $ATHENS_STORAGE:/var/lib/athens \
      -v $ATHENS_BZR:/root/.bazaar \
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
    $env:ATHENS_BZR = "$(Join-Path $pwd .bazaar)"
    md -Path $env:ATHENS_STORAGE
    docker run -d -v "$($env:ATHENS-STORAGE):/var/lib/athens" `
      -v "$($env:ATHENS-BZR):/root/.bazaar" `
      -e ATHENS_DISK_STORAGE_ROOT=/var/lib/athens `
      -e ATHENS_STORAGE_TYPE=disk `
      --name athens-proxy `
      --restart always `
      -p 3000:3000 `
      gomods/proxy:latest
    ```

