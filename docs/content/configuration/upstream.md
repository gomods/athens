---
title: Configuring Upstream Proxy to use Go modules repository
description: How to Fetch modules directly from a Go modules repository such as GoCenter
weight: 1
---

The upstream proxy used when fetching modules directly is by default the actual source, e.g github.com. This can be configured to use a Go modules repository like GoCenter.

1. Create a filter file (e.g /usr/local/lib/FilterForGoCenter) with letter `D` (stands for "direct acccess") in first line. For more details, please refer to documentation on  - [Filtering Modules](/configuration/filter)

    ```
    # FilterFile for fetching modules directly from upstream
    D
    ```
1. If you are not using a config file, create a new config file (based on the sample config.dev.toml and edit values to match your environment).
Additionally in the current or new config file, set the following parameters as suggested:

    ```
    FilterFile = "/usr/local/lib/FilterForGoCenter"
    GlobalEndpoint = "https://<url_to_uptream>"
    # To use GoCenter for example, replace <url_to_upstream> with gocenter.io
    ```
1. Restart Athens specifying the updated current or new config file.

    ```
    /proxy  -config_file <path-to updated  current or new configfile>
    ```
1. Verify the new configuration using the steps mentioned in ("Try out Athens" document)[/try-out], and go through the same walkthrough example.
