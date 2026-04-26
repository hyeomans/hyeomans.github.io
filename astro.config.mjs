// @ts-check

import mdx from "@astrojs/mdx";
import sitemap from "@astrojs/sitemap";
import { defineConfig } from "astro/config";
import rehypeMermaid from "rehype-mermaid";

import tailwindcss from "@tailwindcss/vite";

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

  integrations: [mdx(), sitemap()],

  markdown: {
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
