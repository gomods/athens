---
title: Managing private repos with .netrc files
description: Authenticate athens against private repos
weight: 10
---

## Authenticate private repositories via .netrc

1. Create a .netrc file that looks like the following:

	`machine <ip or fqdn>`

  	`login <username>`
	
  	`password <user password>`

2. Tell Athens through an environment variable the location of that file

	`ATHENS_NETRC_PATH=<location/to/.netrc>`

3. Athens will copy the file into the home directory and override whatever .netrc file is in home directory. Alternatively, if the host of the Athens server already has a .netrc file in the home directory, then authentication should work out of the box.

## Authenticate Mercurial private repositories via .hgrc

1. Create a .hgrc file with authentication data

2. Tell Athens through an environment variable the location of that file

	`ATHENS_HGRC_PATH=<location/to/.hgrc>`

3. Athens will copy the file into the home directory and override whatever .hgrc file is in home directory. Alternatively, if the host of the Athens server already has a .hgrc file in the home directory, then authentication should work out of the box.

