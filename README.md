# Astro Starter Kit: Blog

```sh
bun create astro@latest -- --template blog
```

> 🧑‍🚀 **Seasoned astronaut?** Delete this file. Have fun!

Features:

- ✅ Minimal styling (make it your own!)
- ✅ 100/100 Lighthouse performance
- ✅ SEO-friendly with canonical URLs and OpenGraph data
- ✅ Sitemap support
- ✅ RSS Feed support
- ✅ Markdown & MDX support

## 🚀 Project Structure

Inside of your Astro project, you'll see the following folders and files:

```text
├── public/
├── src/
│   ├── components/
│   ├── content/
│   ├── layouts/
│   └── pages/
├── astro.config.mjs
├── README.md
├── package.json
└── tsconfig.json
```

Astro looks for `.astro` or `.md` files in the `src/pages/` directory. Each page is exposed as a route based on its file name.

There's nothing special about `src/components/`, but that's where we like to put any Astro/React/Vue/Svelte/Preact components.

The `src/content/` directory contains "collections" of related Markdown and MDX documents. Use `getCollection()` to retrieve posts from `src/content/blog/`, and type-check your frontmatter using an optional schema. See [Astro's Content Collections docs](https://docs.astro.build/en/guides/content-collections/) to learn more.

Any static assets, like images, can be placed in the `public/` directory.

## 🧞 Commands

All commands are run from the root of the project, from a terminal:

| Command                   | Action                                           |
| :------------------------ | :----------------------------------------------- |
| `bun install`             | Installs dependencies                            |
| `bun dev`             | Starts local dev server at `localhost:4321`      |
| `bun build`           | Build your production site to `./dist/`          |
| `bun preview`         | Preview your build locally, before deploying     |
| `bun astro ...`       | Run CLI commands like `astro add`, `astro check` |
| `bun astro -- --help` | Get help using the Astro CLI                     |

## Digital wallet contact card

`/card/` is a deliberately minimal contact card for NFC and QR sharing. It exposes only `me@hyeomans.com` and the LinkedIn profile. Generate its deterministic QR and Wallet artwork with:

```sh
npm run card:assets
```

Signed Apple Wallet and Google Wallet artifacts are generated during deployment. Configure these GitHub Actions secrets:

- `APPLE_PASS_TYPE_IDENTIFIER` (for example, `pass.com.hyeomans.contact`)
- `APPLE_TEAM_IDENTIFIER`
- `APPLE_WWDR_CERTIFICATE` (PEM text or base64-encoded PEM)
- `APPLE_SIGNER_CERTIFICATE` (PEM text or base64-encoded PEM)
- `APPLE_SIGNER_KEY` (PEM text or base64-encoded PEM)
- `APPLE_SIGNER_KEY_PASSPHRASE` (only when the key is encrypted)
- `GOOGLE_WALLET_ISSUER_ID`
- `GOOGLE_WALLET_CREDENTIALS` (service-account JSON or its base64 encoding)

`npm run wallet:generate` allows unconfigured local builds and marks unavailable passes honestly. `npm run wallet:verify` is the strict release gate and requires both issuer configurations. No signing secret is committed or delivered to the browser.

`npm run wallet:test` exercises both signing implementations with disposable local certificates and then removes the generated test artifacts.

## 👀 Want to learn more?

Check out [our documentation](https://docs.astro.build) or jump into our [Discord server](https://astro.build/chat).

## Credit

This theme is based off of the lovely [Bear Blog](https://github.com/HermanMartinus/bearblog/).
