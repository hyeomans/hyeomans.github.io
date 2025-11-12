---
title: "The Best Code I Wrote Was The Code I Refactored"
description: "My 'simple' integration with an email marketing software turned into a complex backend monster. Here’s the story of how a split-brain unsubscribe problem forced me to refactor, delete most of my code, and build a much smarter, simpler solution."
pubDate: 2025-11-12T16:48:30+00:00
author: "Hector Yeomans"
tags: ["email marketing software", " loops.so", " api integration", " refactoring"]
lang: "en"
draft: false
heroImage: "./hero.gif"
heroAlt: "Hero image for The Best Code I Wrote Was The Code I Refactored"
---

Last week I dove headfirst into integrating [`loops.so`](https://loops.so/) into my project. If you're not familiar, Loops is a modern email marketing software designed specifically for developers. Choosing the right email marketing software is one challenge; integrating it in a smart, maintainable way is another. **I learned that the hard way**.

My goal seemed *simple enough*: I have a few landing pages offering free resources. When a user subscribes, I want to send them the correct resource. But, as projects do, the requirements list grew into something much more complex.

## What Was I Really Trying To Accomplish?

My initial idea of just "sending an email" quickly morphed into a more detailed set of rules that any robust system should handle. I needed to:

1. Trigger a specific welcome email based on which form the user submitted.
1. **Implement a full double opt-in flow**. The user must confirm their email before getting anything.
1. The confirmation link needed a **unique token that expired in 24 hours** to be secure.
1. Only after a successful confirmation should the system send the correct free resource.
1. Keep a permanent record of which user has received which resource.
1. Prevent a user from getting the same resource email again if they signed up twice.


The core challenge exploded. This wasn't just state management; it was a secure, time-sensitive handshake followed by a long-term user journey. These are the exact problems that good **email marketing software** is designed to solve.

## How I Learned It (By Building A Monster)


My first instinct, as a backend developer, was to control everything. "**The backend is the source of truth**," I told myself, and started architecting my own solution.

The user flow was already getting hairy. A submission would hit an API endpoint, generate a token, save it to a database, and trigger a confirmation email. A second endpoint would validate the token, update my own tables, and then trigger the *real* resource email.

It was a lot, but it seemed manageable. Then I asked myself the killer question: 

> "What happens when a user clicks 'unsubscribe' in an email sent by Loops?"

That's when the whole house of cards started to teeter.


Loops would, of course, handle the unsubscribe. But my backend wouldn't know about it. My `newsletter_subscriptions` table would still show them as an active subscriber. I'd have a classic "split-brain" problem. This is a tell-tale sign that you're trying to manage subscriber state outside of your dedicated email marketing software. The only fix would be to build another endpoint to receive webhooks from Loops just to sync unsubscribe statuses.

I stopped right there. I was accidentally architecting my own, single-purpose, and vastly inferior email marketing software from scratch. The complexity was spiraling out of control.

The problem wasn't my desire for a backend; it was the scope of its responsibility. I needed to let the tools do what they're best at.

## The Real "Aha!" Moment: Reading The Docs

Just as I was about to start coding this brittle system, I decided to take one last, thorough look through the Loops documentation. And there it was, plain as day: [Forms](https://loops.so/docs/forms/custom-form).

Loops doesn't just give you an API; it gives you hosted forms that have double opt-in built right in. The entire handshake, the secure link, the confirmation tracking... it was all a feature, waiting to be used.

My beautiful, complex whiteboard architecture was completely unnecessary.


I erased the whiteboard. The entire backend plan was gone. My backend's involvement in this entire flow went from "complex" to "zero."

Here is the new, ridiculously simple workflow:

1. [Create a Form in Loops](https://loops.so/docs/forms/custom-form). I made one for each free resource. Inside the form settings, I just checked a box: "Send double opt-in."
1. **Embed the Form**. Loops provides a simple HTML snippet. I pasted it onto my landing page.
1. **Create a Workflow**. Inside Loops, I created a "Loop" with a simple rule:
  - Trigger: When a contact confirms their subscription to the 'Figma Cheatsheet' Form.
  - Action: Send the email containing the Figma Cheatsheet.


That's it. The entire process, from submission to confirmation to resource delivery, is handled by Loops. There's no split-brain. Unsubscribes are managed in one place. And I didn't have to write a single line of backend logic for it.

## Why This Matters (And What's Next)

The lesson wasn't just to use a tool's features. It was a potent reminder to fight the developer instinct to build everything yourself. My initial approach failed because I saw Loops as just an API, a tool to be commanded by my "smarter" backend. I was wrong. The value of a great email marketing software isn't just the API; it's the complete solutions that save you from writing code in the first place.

Now, this embedded form solution is perfect for getting started quickly and is incredibly robust. But what happens when you have a highly-styled, custom React form that's deeply integrated with your site's design? Do you have to go back to building that complex backend to handle the submission?

The answer is no. And that's what I'll cover next.

In my next post, I'll show you how to have the best of both worlds: connecting your own custom frontend form directly to the Loops API. We'll build a solution that gives you full design control while still letting Loops handle the heavy lifting of the double opt-in and the email sequences. Stay tuned.