---
title: Using an upstream Go modules repository (deprecated)
description: How to Configure Athens to Fetch Missing Modules From an Upstream Module Repository Like GoCenter, or Another Athens Server
weight: 7
---

>Note: the filter file that this page documents is deprecated. Please instead see ["Filtering with the download mode file"](./download) for updated instructions on how to set upstream repositories in Athens.

By default, Athens fetches module code from an upstream version control system (VCS) like github.com, but this can be configured to use a Go modules repository like GoCenter or another Athens Server.

1. Create a filter file (e.g ```/usr/local/lib/FilterForGoCenter```) with letter `D` (stands for "direct access") in first line. For more details, please refer to documentation on  - [Filtering Modules](/configuration/filter)

    ```
    # FilterFile for fetching modules directly from upstream
    D
    ```
2. If you are not using a config file, create a new config file (based on the sample config.dev.toml) and edit values to match your environment).
Additionally in the current or new config file, set the following parameters as suggested:

    ```
    FilterFile = "/usr/local/lib/FilterForGoCenter"
    GlobalEndpoint = "https://<url_to_upstream>"
    # To use GoCenter for example, replace <url_to_upstream> with gocenter.io
    # You can also use https://proxy.golang.org to use the Go Module mirror
    ```
1. Restart Athens specifying the updated current or new config file.

    ```
     <path_to_athens>/proxy  -config_file <path-to updated  current or new configfile>
    ```
1. Verify the new configuration using the steps mentioned in ["Try out Athens" document](/try-out), and go through the same walkthrough example.
