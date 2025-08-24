# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Hugo-based static site blog (hyeomans.com) featuring bilingual content (English/Spanish) with a focus on technology, open data, transparency, and open government. The site uses the PaperMod theme with chart functionality.

## Development Commands

### Local Development
```bash
# Initialize submodules after cloning
git submodule update --init --recursive

# Start development server
hugo server

# Start with drafts and future posts
hugo server -D -F

# Build for production
hugo --gc --minify
```

### Content Management
```bash
# Create new post
hugo new posts/post-name.md

# Create new Spanish post
hugo new content/posts/post-name.es.md
```

## Architecture

### Site Structure
- **Bilingual Setup**: English (`en`) and Spanish (`es`) languages configured
- **Content Organization**: Posts in `/content/posts/` with language-specific files using `.es.md` suffix
- **Themes**: Uses both `hugo-chart` and `PaperMod` themes (PaperMod is primary)
- **Static Assets**: Images and resources in `/static/img/`
- **Output**: Generated site in `/public/` directory

### Key Configuration (config.yaml)
- **Base URL**: https://hyeomans.com/
- **Themes**: `["hugo-chart", "PaperMod"]` - enables chart shortcodes with PaperMod styling
- **Search**: Fuse.js-powered search functionality enabled
- **Features**: Code copy buttons, reading time, share buttons, breadcrumbs, TOC

### Content Types
- **Blog Posts**: Technical content, government transparency analysis, data visualization
- **Languages**: Content available in both English and Spanish
- **Special Features**: Chart integration for data visualization posts

## Deployment

- **Platform**: GitHub Pages
- **Automation**: GitHub Actions workflow (`.github/workflows/hugo.yaml`)
- **Hugo Version**: 0.139.3 (specified in workflow)
- **Build Process**: Automatically builds and deploys on push to main branch

## Content Guidelines

### Post Structure
- Posts support both languages with `.es.md` suffix for Spanish content
- Hero images should be placed in post-specific subdirectories
- Chart shortcodes available via hugo-chart theme
- Front matter should include proper taxonomies (tags, series)

### Asset Management
- Images go in `/static/img/`
- Post-specific assets can be co-located with content in subdirectories
- Charts and data visualizations are common content types

## Development Notes

- **Theme Customization**: PaperMod theme is included as a submodule
- **Multilingual**: Full i18n support with language-specific menus and content
- **Analytics**: Google Analytics configured (G-CQCQ4X7JHE)
- **SEO**: Comprehensive meta tags, OpenGraph, and structured data