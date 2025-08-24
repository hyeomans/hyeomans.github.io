---
title: "Image Test Post"
date: 2025-08-24T09:00:03-07:00
draft: true

showToc: true
TocOpen: false
hidemeta: false

disableShare: false
disableHLJS: false
hideSummary: false
searchHidden: false
ShowReadingTime: true
ShowBreadCrumbs: true
cover:
  image: "hero.jpg"
  alt: "Hero image for testing page bundles"
  caption: "This is a hero image from the same folder as the post"
  relative: true # when using page bundles set this to true
  hidden: false # only hide on current single page
---

This is a test post to verify that images work correctly with Hugo page bundles.

**Note:** The hero image appears above this text because your site config has `hiddenInSingle: true`, so cover images in front matter don't show on individual posts. Instead, we use `![](./hero.jpg)` in the content like your other posts.

## Inline Image Test

Here's an inline image referenced from the same folder:

![Inline test image](inline-image.jpg)

## Image Features

- **Hero Image**: Set in the front matter, appears at the top
- **Inline Images**: Referenced directly by filename since they're in the same folder
- **Page Bundle**: All images are co-located with the markdown file

## File Structure

```
content/posts/image-test-post/
├── index.md (this file)
├── hero.jpg (cover image)
└── inline-image.jpg (referenced in content)
```

Both images should render correctly when you run `make serve`.
