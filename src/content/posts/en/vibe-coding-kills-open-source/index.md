---
title: "Does Vibe Coding Kill Open Source? Notes on a New Paper"
description: "A plain-English breakdown of \"Vibe Coding Kills Open Source\" and practical ways maintainers, users, and AI toolmakers can keep OSS sustainable."
pubDate: 2026-02-05T14:26:20+00:00
author: "Hector Yeomans"
tags: ["open source", "vibe coding", "generative ai", "software economics"]
lang: "en"
draft: false
heroImage: "./hero.gif"
heroAlt: "Crafting a vibe-coded app."
---

Recently I started building a Go API and did what I usually do first: signup/signin. That quickly turned into the full email/password gauntlet: forgot password, email verification, resets, edge cases, the whole thing.

I set the foundation the way I've done it for years: an OpenAPI-first contract, Docker + testcontainers, and a layered architecture that keeps the seams clear. Once that was in place, I switched into my `ralph-wiggum` mode: ship a feature, verify it, repeat.

And that's when it hit me: maintaining a correct, secure email/password system as a solo developer is a time sink I didn't want to sign up for. So I ripped it out and kept only social login. There's a demo in the hero section showing how it works.

That experience is why I've been thinking so much about "vibe coding": shipping real features by collaborating with an AI agent instead of writing every line yourself. It's legitimately useful. It also changes how we touch the upstream projects we rely on.

Then I read a recent economics paper with a blunt title: "Vibe Coding Kills Open Source." The claim is simple: if agents sit between developers and the projects they depend on, open source can get less sustainable, not because usage drops, but because the interactions maintainers rely on (docs visits, bug reports, Q&A, contributions) get routed somewhere else.

Paper: [Vibe Coding Kills Open Source (Koren, Békés, Hinz, Lohmann)](https://arxiv.org/abs/2502.06059)

## TL;DR

- Vibe coding makes reuse cheaper, so more software gets built.
- It can also reduce human touchpoints with upstream projects (docs, issues, Q&A).
- If money still follows those touchpoints, maintainers can lose income even while usage grows.
- The fix isn't "stop using AI." It's better attribution and easier ways to pay maintainers.

## What the paper means by "vibe coding"

In the paper's framing, vibe coding is when an AI agent:

- Picks libraries/packages for you
- Assembles them into a working app
- Iterates on code via prompts
- Often shields you from the upstream ecosystem (docs, changelogs, issue trackers)

That last part is the key. Vibe coding doesn't just change how code is written. It changes how demand shows up for open source.

## The paper's core argument: more output, fewer touchpoints

The paper describes two effects pulling in opposite directions:

1. Building gets cheaper, so more people build. More OSS gets used.
2. When an agent mediates usage, fewer humans read docs, ask questions publicly, file issues, or become contributors.

If maintainers mostly get paid through public visibility and interaction, vibe coding can shrink the return to maintaining OSS even if usage grows.

## Why those touchpoints matter

In practice, most maintainers don't get paid per download. They get paid through a messy bundle of signals and relationships:

- Someone finds the project via docs/search, then hires the maintainer
- A company becomes dependent on a library, then pays for support
- Sponsors show up because the project looks active and visible
- Contributors arrive through issues/discussions and reduce maintainer burden

If AI tools route people around those surfaces, the maintainer's funnel gets weaker.

## My take: mostly right, with a few missing pieces

I think the paper is right about the basic risk: we're changing the incentive structure without replacing it.

But I'd add a few nuances:

- The touchpoints don't vanish. They move, often into private chats, telemetry, or vendor-run systems. When that happens, they're less useful to everyone else.
- Some projects already make money without much public Q&A (support contracts, hosted offerings, dual licensing). Those models may hold up better.
- The biggest risk isn't "AI writes code." It's unknown dependencies and low reciprocity: users don't know what they used, and maintainers never learn who depends on them.

## What would help (so OSS doesn't get quietly hollowed out)

Some practical changes would go a long way.

### 1) Make dependency attribution the default

AI tools should produce a dependency list (an SBOM) as a first-class artifact:

- What OSS projects were used (direct + transitive)
- Links to upstream docs and changelogs
- A short "how to report bugs responsibly" note

That turns invisible usage into something maintainers can actually see.

### 2) Make it easy to pay without "engagement theater"

If maintainers only get paid when a human shows up in the right place, that's fragile.

Better options:

- Org-level sponsorship budgets tied to dependency audits
- Support contracts for critical packages
- Funds that distribute money based on dependency graphs
- AI tools sharing revenue with the OSS they lean on

### 3) Treat maintenance as a supply chain problem

If your business depends on OSS, you're already in the software supply chain. The right mental model is:

> "We pay to keep critical infrastructure healthy."

Not:

> "We'll pay if the maintainer becomes famous enough."

## What you can do today (if you vibe code)

If you're shipping with AI agents, here's a practical checklist:

1. **Ask your agent for a dependency list** before you merge.
2. **Read the upstream docs for the top few dependencies** you're leaning on most.
3. If you hit a bug, **file it upstream** (with a minimal repro) instead of leaving it trapped in a private chat.
4. If a project saved you real time, **sponsor it**. Even a small recurring amount helps.
5. If you're at a company: keep a short "top dependencies" list and assign budget + ownership.

## If you maintain OSS: reduce friction for the AI era

You shouldn't have to do extra work to survive, but a few small changes can help:

- Add `FUNDING.yml` and a short "Support this project" section in the README
- Add a "Bug reports" section that explains what you need (versions, repro, logs)
- Publish clear release notes and upgrade guides
- Consider a paid support path for orgs that need guarantees

## FAQ

### Is the solution to stop using AI coding tools?

No. The goal is to make sure value flows back upstream when AI tools accelerate downstream software creation.

### Won't more usage naturally lead to more contributions?

Sometimes, yes. But the paper's point is that usage mediated by agents can increase without increasing the interactions that turn users into contributors or sponsors.

### What should AI toolmakers do?

At minimum: show attribution and links by default. Better: route money and high-quality bug reports back to upstream projects.

## Further reading

- The paper: [Vibe Coding Kills Open Source](https://arxiv.org/abs/2502.06059)
- GitHub Sponsors: [GitHub Sponsors](https://github.com/sponsors)
- Open Collective: [Open Collective](https://opencollective.com/)
