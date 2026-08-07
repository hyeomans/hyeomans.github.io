# Digital Wallet Contact Card

## Outcome

Publish a privacy-minimal contact card at `https://hyeomans.com/card/`. The page makes Apple Wallet and Google Wallet the primary save actions and exposes only Hector Yeomans's public email address and LinkedIn profile. The same page URL is the target for the on-page QR code and for a physical NFC tag.

## Chosen approach

Use pre-generated, signed wallet artifacts with the existing static Astro/GitHub Pages deployment.

- Apple Wallet receives a signed generic `.pkpass` generated during deployment.
- Google Wallet receives a signed `pay.google.com/gp/v/save/…` URL generated during deployment.
- Signing material stays in GitHub Actions secrets and is never sent to the browser or committed.
- The public page contains no analytics, structured personal metadata, phone number, location, form, third-party embed, or third-party JavaScript.
- If either issuer is not configured, its action remains visibly unavailable instead of linking to a broken or unsigned pass.

This is preferable to a dynamic pass server because the card has one fixed identity, does not need pass updates or per-recipient identifiers, and is hosted on a static site. It is preferable to a commercial pass provider because a provider would add another party to the contact-sharing flow.

## Public payload

Every surface uses the same allowlist:

- Name: Hector Yeomans
- Role: Staff Software Engineer
- Email: `me@hyeomans.com`
- LinkedIn: `https://www.linkedin.com/in/hector-yeomans/`
- Card URL: `https://hyeomans.com/card/`

No phone field is present. The QR code contains only the card URL rather than an inline vCard, so future content changes do not require printing a new code or reprogramming an NFC tag.

## Components and data flow

1. `scripts/generate-card-assets.mjs` produces a deterministic QR code and Wallet-safe PNG artwork.
2. `scripts/generate-wallet-passes.mjs` reads issuer credentials only from environment variables, signs the two pass formats, and writes public deployment artifacts.
3. `src/data/generated-wallet-links.json` records which passes were generated and supplies the Google save URL to Astro at build time.
4. `src/pages/card/index.astro` renders a standalone, no-index contact card without the site's analytics-enabled layout.
5. The GitHub Pages workflow runs asset and pass generation before `astro build`.
6. Apple signing uses the system OpenSSL and ZIP tools, minimizing the build dependency surface. The Google JWT is signed and immediately verified with Node's cryptography API.
7. `scripts/verify-card-privacy.mjs` inspects the built page and fails if it finds telephone fields, unrelated social profiles, analytics, or missing allowlisted contact details.
8. A post-deployment check verifies the live card and, once Apple signing is configured, requires GitHub Pages to serve the pass with `application/vnd.apple.pkpass`.

The production NFC check covers the web side of an NFC handoff: the exact URL stored on the tag must resolve to the verified card. The physical tag and reader still require an on-device tap test. The printed/on-screen QR encodes the same URL and is the fallback if a particular phone cannot read the tag.

## Failure behavior

Issuer credentials are external prerequisites. A normal local or Pages build still succeeds when they are absent, but the corresponding Wallet control is marked unavailable. `npm run wallet:verify` is the release gate that requires both signed outputs. Invalid credentials fail the signing step without leaking secret contents.

## Verification

- Run the production Astro build.
- Run the privacy allowlist check against `dist/card/index.html`.
- Inspect the page at mobile and desktop sizes.
- When credentials exist, verify the Apple archive manifest/signature and the Google JWT structure, then test each save action on a real device.
