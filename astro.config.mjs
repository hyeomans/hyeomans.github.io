// @ts-check

import mdx from "@astrojs/mdx";
import sitemap from "@astrojs/sitemap";
import { defineConfig } from "astro/config";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";
import rehypeMermaid from "rehype-mermaid";

import tailwindcss from "@tailwindcss/vite";

const slugifyTag = (value) =>
  value
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "")
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");

const walkMarkdownFiles = (dir) => {
  const entries = readdirSync(dir);
  const files = [];

  for (const entry of entries) {
    const path = join(dir, entry);
    const stats = statSync(path);
    if (stats.isDirectory()) {
      files.push(...walkMarkdownFiles(path));
    } else if (entry.endsWith(".md") || entry.endsWith(".mdx")) {
      files.push(path);
    }
  }

  return files;
};

const getIndexableTagSlugs = () => {
  const countsByLang = {
    en: new Map(),
    es: new Map(),
  };

  for (const file of walkMarkdownFiles(join(process.cwd(), "src/content/posts"))) {
    const source = readFileSync(file, "utf8");
    const frontmatter = source.match(/^---\n([\s\S]*?)\n---/)?.[1];
    if (!frontmatter || /(?:^|\n)draft:\s*true(?:\n|$)/.test(frontmatter)) continue;

    const lang = frontmatter.match(/(?:^|\n)lang:\s*["']?(en|es)["']?(?:\n|$)/)?.[1];
    const tagSource = frontmatter.match(/(?:^|\n)tags:\s*\[([^\]]*)\]/)?.[1];
    if (!lang || !tagSource) continue;

    for (const match of tagSource.matchAll(/["']([^"']+)["']/g)) {
      const slug = slugifyTag(match[1].trim());
      if (!slug) continue;
      countsByLang[lang].set(slug, (countsByLang[lang].get(slug) ?? 0) + 1);
    }
  }

  return {
    en: new Set([...countsByLang.en].filter(([, count]) => count > 1).map(([slug]) => slug)),
    es: new Set([...countsByLang.es].filter(([, count]) => count > 1).map(([slug]) => slug)),
  };
};

const indexableTagSlugs = getIndexableTagSlugs();

const shouldIncludeInSitemap = (page) => {
  const { pathname } = new URL(page);
  if (pathname === "/posts/todo/") return false;

  const englishTag = pathname.match(/^\/tags\/([^/]+)\/$/);
  if (englishTag) return indexableTagSlugs.en.has(englishTag[1]);

  const spanishTag = pathname.match(/^\/es\/tags\/([^/]+)\/$/);
  if (spanishTag) return indexableTagSlugs.es.has(spanishTag[1]);

  return true;
};

// Astro highlights fenced code before this reaches the browser. Normalize Mermaid
// fences back to the shape expected by rehype-mermaid's pre-mermaid strategy.
function markMermaidCodeBlocks() {
  return (tree) => {
    const textContent = (node) => {
      if (!node || typeof node !== "object") return undefined;
      if (node.type === "text") return node.value;

      if (Array.isArray(node.children)) {
        return node.children.map((child) => textContent(child) ?? "").join("");
      }

      return "";
    };

    const visit = (node) => {
      if (!node || typeof node !== "object") return;

      if (node.type === "element" && node.tagName === "pre") {
        const properties = node.properties ?? {};
        if (properties.dataLanguage === "mermaid" || properties["data-language"] === "mermaid") {
          node.properties = { className: ["mermaid"] };
          node.children = [{ type: "text", value: textContent(node).trim() }];
        }
      }

      if (Array.isArray(node.children)) {
        for (const child of node.children) visit(child);
      }
    };

    visit(tree);
  };
}

// https://astro.build/config
export default defineConfig({
  site: "https://hyeomans.com",
  base: "/",
  trailingSlash: "always",

  i18n: {
    defaultLocale: "en",
    locales: ["en", "es"],
    routing: {
      prefixDefaultLocale: false,
    },
  },

  integrations: [mdx(), sitemap({ filter: shouldIncludeInSitemap })],

  markdown: {
    shikiConfig: {
      themes: {
        light: "github-light",
        dark: "github-dark",
      },
      defaultColor: false,
      wrap: false,
    },
    rehypePlugins: [markMermaidCodeBlocks, [rehypeMermaid, { strategy: "pre-mermaid" }]],
  },

  image: {
    service: {
      entrypoint: 'astro/assets/services/sharp'
    }
  },

  vite: {
    plugins: [tailwindcss()],
  },
});
