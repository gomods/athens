---
title: "Build-in Shortcodes"
date: 2017-10-17T15:26:15Z
draft: false
weight: 10
description: "calling built-in Shortcodes into your content files."
---

See https://gohugo.io/content-management/shortcodes/#use-hugo-s-built-in-shortcodes

## figure

{{< figure src="/images/pexels-photo-196666.jpeg" title="2 People Sitting With View of Yellow Flowers during Daytime" >}}

## gist

{{< gist spf13 7896402 "img.html" >}}

## highlight

{{< highlight html >}}
<section id="main">
  <div>
   <h1 id="title">{{ .Title }}</h1>
    {{ range .Data.Pages }}
        {{ .Render "summary"}}
    {{ end }}
  </div>
</section>
{{< /highlight >}}

## instagram

{{< instagram BWNjjyYFxVx >}}

## tweet

{{< tweet 877500564405444608 >}}

## vimeo
{{< vimeo 146022717 >}}

## youtube
{{< youtube w7Ft2ymGmfc >}}
