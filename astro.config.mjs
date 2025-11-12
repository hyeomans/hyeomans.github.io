// @ts-check

import mdx from "@astrojs/mdx";
import sitemap from "@astrojs/sitemap";
import { defineConfig } from "astro/config";

import tailwindcss from "@tailwindcss/vite";

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
      strategy: "pathname",
    },
  },

  integrations: [mdx(), sitemap()],

  image: {
    service: {
      entrypoint: 'astro/assets/services/sharp'
    }
  },

  vite: {
    plugins: [tailwindcss()],
  },
});
