---
title: "Software as stitching"
date: 2018-12-27T00:00:00-07:00
updated: 2018-12-27T00:00:00-07:00
aliases: ["/software-as-stitching"]
author: "Hector Yeomans"
description: "Exploring how most software development involves stitching together APIs and third-party services, drawing parallels to the art of sewing."
tags: ["software-development", "api-integration", "programming", "software-architecture"]
ShowReadingTime: true
ShowBreadCrumbs: true
---

![](/img/stitching.jpg)

<sub><sup>Photo by Alexander Andrews on Unsplash</sup></sub>

Lately, I've been thinking that most of the software I've done consist of stitching of API's. Imagine you work on
the team that is in charge of handling payments. It doesn't matter if you're working on the front-end or the back-end,
it usually goes like this:

- We (the company you work for) have a new payment provider.
- We must integrate with their API.
- We must accommodate our business logic to their API.

You have to take consideration of several layers in between, this list includes but is not limited to: validation, logging, error handling, persistence, a source of truth, actual business logic. Almost everything becomes a cross-cutting concern, authentication, logging, caching; however, in the end, the hardest part is making sense or making an abstraction of
the new API to conform to the company business logic.

This is where my idea of stitching comes to mind; you don't know from the beginning how everything has to be laid out. Just as you stitch, you don't know for sure if the combination of API call is going to work as you're picturing it in your head.

![](/img/stitching-2.jpg)
<sub><sup>Photo by Dương Trần Quốc on Unsplash</sup></sub>

I'm not saying that software stitching is easy, but it can become repetitive.

![](/img/stitching-3.jpg)
