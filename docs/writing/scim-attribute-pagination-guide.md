# Writing workbook: SCIM attribute cursor pagination

This is a working guide for Hector Yeomans to write an original article for InfoQ. It replaces the earlier AI-written article. The questions are for you to answer from your experience and source reading; they are not prompts to generate the article with AI.

Aim for 3,000–4,000 words in your finished article. The section budgets below total 3,450 words, leaving room for editing and small examples. This workbook's instructions and your research notes do not count toward that total.

## Start here: one experience, in your own words

Spend 15–20 minutes answering these questions in rough notes. Sentence fragments are fine. Start with what you remember, then check records before turning memories into factual claims.

1. What SCIM work have you personally done: service-provider implementation, connector development, SDK work, integration testing, or operations?
2. What exact task was your system trying to complete when a large attribute became a problem?
3. Which resource and attribute were involved? Where were their values stored or derived?
4. What did you observe: a timeout, memory spike, slow query, truncated result, incorrect reconciliation, or something else?
5. What evidence located the problem? Describe the trace, query plan, test, incident record, or code path.
6. What did you initially believe was wrong? What evidence changed your understanding?
7. What workaround did you choose, and what cost or limitation did it leave?
8. Which part of this proposal addresses that experience? Which part would not have helped?
9. What would you tell another implementer before they attempted the same integration?
10. What can you share publicly, and what needs anonymization or omission?

**Your notes:**

_Write here._

**Ready to proceed when:** you have one concrete experience, know which details you can substantiate, and can explain your personal role. You do not need a dramatic outage or a successful deployment of this draft.

If you have relevant SCIM experience but have not implemented the proposal, say so. You can evaluate a proposal against a real constraint without claiming that you deployed it. If you lack a relevant case, choose a reproducible experiment you can perform and describe it as an experiment. Leave unperformed tests and unmeasured results explicitly unresolved.

## Decide what you want the reader to learn

Pick the emphasis your evidence supports:

- **Provider implementation:** where work must be bounded, what changes in the storage path, and what a small response fails to prove.
- **Connector correctness:** how partial membership affects local state, retries, reconciliation, and subsequent writes.
- **Architecture and adoption:** whether this mechanism fits an existing integration and what must be negotiated before rollout.

The outline below uses provider implementation as its main thread and connects it to client correctness. Shift space toward your strongest experience. You do not have to give every topic equal weight.

Answer before drafting:

- Which reader is this for, and what decision are they facing?
- What specific thing will they understand after reading that they probably do not understand now?
- What is your most useful insight beyond what a reader can obtain directly from the draft?
- What is the strongest objection to your argument?
- What evidence would make you reduce your confidence in the proposal's impact?

**Your working thesis:**

_Write one sentence in your own words. Make the claim as strong as the evidence allows._

**Your scope:**

_Write what this article will cover and which adjacent topics it will leave out._

## Keep an evidence sheet

For every consequential claim, record which kind of statement you are making. These labels belong in your notes; the article can communicate the distinctions through ordinary prose.

| Kind | What to record | How to use it |
| --- | --- | --- |
| Personal experience | Your role, observation, date or period, and supporting record | Describe what happened within that scope. |
| Experiment | Setup, data, procedure, result, and limitations | State that it was tested, and under what conditions. |
| Specification fact | Document revision, section, and relevant requirement | Explain the rule and link to its source. |
| Your judgment | Reasoning, assumptions, alternatives, and uncertainty | Own the recommendation or prediction. |
| Illustrative scenario | Why the example is useful and which numbers are invented | Label it hypothetical; keep it separate from results. |

**Claim to substantiate:**

**Kind:**

**Evidence or source:**

**What the evidence does not establish:**

Use as many entries as you need. An anonymized case can be useful without disclosing a customer. If you cannot verify a number, omit it or identify it as an estimate with a defensible basis. The earlier draft's two-million-member example was hypothetical; it is not evidence about your work.

## Article plan and word budget

These are working labels, not mandatory published headings. Write the technical sections first, then the opening, ending, and takeaways.

| Article component | Words |
| --- | ---: |
| Five key takeaways, written last | 150 |
| 1. The implementation problem you encountered | 250 |
| 2. The gap ordinary pagination leaves | 300 |
| 3. Follow one resource through the proposal | 400 |
| 4. Make the database work match the page boundary | 500 |
| 5. Preserve partial state and safe updates | 400 |
| 6. Define what completion means under change | 350 |
| 7. Introduce the behavior to existing clients | 350 |
| 8. Account for security and total workload | 300 |
| 9. Explain the impact and the next decision | 450 |
| **Total** | **3,450** |

For each section, answer the questions as notes, check the cited material, and then write connected prose yourself. Choose the questions that advance your argument; this is not a requirement to publish a question-and-answer article.

## 1. The implementation problem you encountered — 250 words

**Purpose:** give the reader a concrete reason to care and establish what you know firsthand.

**Prompts:**

- What operation did the client perform, and what did it reasonably expect back?
- What was the first symptom visible to you?
- What detail made this more than an ordinary slow endpoint?
- Was the expensive part finding parents, finding relationships, loading referenced objects, or serializing values?
- What assumption in the implementation stopped holding as the attribute grew?
- What was your responsibility in investigating or fixing it?
- Where can you introduce the proposal as a possible response to this problem without implying adoption or proven results?

**Bring:** one shareable incident detail, code observation, or measured result. A small HTTP request can establish the trigger if it helps.

**Write:** two or three paragraphs moving from the operation to its consequence and the question you want to examine. Give the reader the central issue early; suspense is unnecessary.

**Done when:** a colleague can name the operation, the failure or constraint, and your role. The opening contains no invented personal history.

**Your section:**

_Write here._

## 2. The gap ordinary pagination leaves — 300 words

**Purpose:** explain the technical distinction that makes the proposal relevant.

**Prompts:**

- In your implementation, what exactly does a resource-level page size limit?
- How could a request containing one parent still cause excessive work?
- What does the current schema lead the client to expect for the attribute?
- How do selecting, excluding, and silently truncating values differ in their effect on a client?
- Which existing workaround did you use or investigate?
- Would inverse filtering or a separate relationship endpoint solve your use case? What dependency or limitation would remain?
- Which gap does the proposed behavior address beyond collection cursor pagination?

**Read:** [RFC 7643, Group](https://www.rfc-editor.org/rfc/rfc7643.html#section-4.2), [RFC 7644, attribute selection](https://www.rfc-editor.org/rfc/rfc7644.html#section-3.9), [RFC 9865](https://www.rfc-editor.org/rfc/rfc9865.html), and [draft Section 1](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-1).

**Write:** explain the two page boundaries using the resource from your opening. Assume the reader knows HTTP APIs and pagination. Introduce SCIM-specific semantics where they affect the argument.

**Done when:** the reader understands why reducing the parent page size may leave your problem unsolved. Existing alternatives receive a fair account.

**Your section:**

_Write here._

## 3. Follow one resource through the proposal — 400 words

**Purpose:** teach the protocol through one coherent exchange that you have checked.

**Prompts:**

- How would your client discover support, and how would deferred behavior be established for it?
- Which endpoint returns parents, and which request retrieves values of the chosen attribute?
- What does the client learn from a deferred descriptor?
- What is different about an absent array with deferral metadata and an empty completed result?
- How does the client get the first page and request the next one?
- Which parameters and selected subfields must stay consistent during continuation?
- How does it know when a normal traversal has ended?
- What changes if a request selects several protected attributes?
- What special handling does a zero-count request require?

**Read:** [draft Sections 4–8](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-4) and [examples in Section 9](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-9).

**Bring:** your own minimal request/response sequence, checked against the pinned revision. If it comes from a prototype, identify that. If it is illustrative, label it accordingly.

**Write:** explain the state change after each important exchange. Include only fields needed to understand it, and label abbreviated payloads. Keep identifiers and projections consistent across examples.

**Done when:** a reader can follow discovery, deferral, first page, continuation, and completion without inventing a missing step. Examples distinguish root and attribute pagination.

**Your section:**

_Write here._

## 4. Make the database work match the page boundary — 500 words

**Purpose:** contribute implementation depth that a protocol summary cannot provide.

**Prompts:**

- Trace the request from handler to storage. Where does the implementation first discover member values?
- Does a limit apply before or after relationship identifiers and referenced objects are loaded?
- Which eager-loading behavior, join, sort, or count could still perform unbounded work?
- What ordering key would you use? How would you handle ties and mutable sort values?
- What index would support the actual predicates, including tenant boundaries and authorization?
- How would look-ahead work, and when would the extra candidate leave the loading pipeline?
- Could reference resolution create one query per returned value?
- What happens when a byte budget is reached before the requested count?
- What happens if authorization filters out most candidates?
- What would convince you that database work improved, rather than only response size?
- Which tradeoff would you accept, and what would you measure before accepting it?

**Read:** [draft Section 8](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-8), [ordering in Section 10](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-10), and [implementation guidance in Appendix C](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#appendix-C).

**Bring:** a query plan, small query or pseudocode example, or a precise account of a code path you inspected. Use only material you understand and can explain.

If you run a comparison, record dataset shape, index definitions, cache conditions, concurrency, measurement method, and number of runs. Compare equivalent work. Report response size separately from rows examined, allocations, and elapsed time. Label suggested measurements as a plan until you perform them.

**Write:** organize this around one implementation choice and its consequences. Give the other details only enough space to show what could invalidate that choice.

**Done when:** the section identifies the earliest useful work boundary, supports the proposed storage approach, and states what remains untested.

**Your section:**

_Write here._

## 5. Preserve partial state and safe updates — 400 words

**Purpose:** explain how a read optimization can affect identity correctness.

**Prompts:**

- What does your client do with a missing attribute today?
- Does its model retain extension metadata or discard unknown fields?
- Where could incomplete membership become an apparently complete collection?
- Does any workflow read a resource, change one field, and save the whole object?
- Follow that workflow with one membership page: what exact write could result?
- How could the client express a targeted change instead?
- What can the provider detect, and what information might the client remove?
- How would you represent deferred, partly consumed, and completed data?
- Which test proves that changing an unrelated field cannot remove unseen membership?
- If this is a risk you identified rather than a failure you observed, how will you say that?

**Read:** [RFC 7644, replacement](https://www.rfc-editor.org/rfc/rfc7644.html#section-3.5.1), [PATCH](https://www.rfc-editor.org/rfc/rfc7644.html#section-3.5.2), and [draft mutation safety](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-12).

**Write:** walk through one concrete read-to-write path. Explain a design change in your own terms, including why it belongs in the client, provider, or both.

**Done when:** the reader can identify the unsafe assumption, the possible consequence, and a testable way to avoid it. You do not imply that metadata alone guarantees safety.

**Your section:**

_Write here._

## 6. Define what completion means under change — 350 words

**Purpose:** make the consistency requirement follow from the business operation.

**Prompts:**

- What action will your application take after it believes it has all values?
- Does that action require a current view, a snapshot, or eventual convergence?
- What happens if a value is inserted before the cursor boundary?
- What happens if a member disappears or its ordering key changes?
- How could a crash between page processing and checkpoint storage cause repetition?
- What recovery would you choose when a cursor expires?
- What does an attribute version establish, and how might it differ from the parent version?
- What guarantee is missing if the application uses absence to remove access?
- Which guarantee must come from another mechanism or from the provider?

**Read:** [draft consistency rules](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-11), [cursor context](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-10), and [errors](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-13).

**Write:** choose one concurrent change and one recovery event. Explain their consequences for your actual workload. Distinguish finishing the sequence from obtaining a snapshot.

**Done when:** you state the required guarantee, what the mechanism provides, and the remaining responsibility of the application.

**Your section:**

_Write here._

## 7. Introduce the behavior to existing clients — 350 words

**Purpose:** give team leads a concrete way to evaluate adoption.

**Prompts:**

- Which clients can you change, and which are outside your control?
- What evidence would show a connector understands partial data?
- Why might capability advertisement be insufficient for an existing integration?
- What scope would you choose for an initial rollout: client, tenant, endpoint, or another boundary?
- Which read and write flows would you replay before enabling the behavior?
- What should happen to a legacy request the provider cannot serve safely?
- What failure would stop expansion of the rollout?
- What is the fallback if legacy full reads are themselves unsafe?
- Who owns connector changes, operational limits, and support documentation?

**Read:** [draft compatibility and discovery](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-4) and [proposed errors](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-13).

**Bring:** a known compatibility constraint or an explicitly proposed test matrix. Attribute vendor-specific behavior to documentation you checked or a named integration version you tested.

**Write:** describe the smallest meaningful rollout and its acceptance criteria. A short sequence of actions will be easier to scan than several general paragraphs.

**Done when:** a reader can identify a first integration, required checks, a stop condition, and a safe response to failed migration.

**Your section:**

_Write here._

## 8. Account for security and total workload — 300 words

**Purpose:** show the operational cost that remains after individual pages are bounded.

**Prompts:**

- What must prevent a cursor from being reused for a different tenant or parent?
- What happens when permission changes between pages?
- Could cursor contents or deferred metadata reveal sensitive information?
- How many follow-up requests does a full traversal create at your chosen page size?
- Which assumptions make that estimate valid?
- How would client concurrency, retries, and server scheduling affect other tenants?
- At what workload would an export or another retrieval path be a better fit?
- What would you measure for the whole synchronization, not just an individual request?

**Read:** [draft cursor integrity](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-10), [operational limits](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-14), and [security](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-15).

**Write:** connect one authorization boundary and one capacity tradeoff to your system. Use transparent arithmetic if useful; label estimates.

**Done when:** the reader understands what work moves to follow-up requests and what controls still have to exist.

**Your section:**

_Write here._

## 9. Explain the impact and the next decision — 450 words

**Purpose:** make a defensible case for attention and leave the reader with an action.

**Prompts:**

- Return to your opening: what would change if the proposed contract were available to both sides?
- Which workaround or provider-specific assumption could disappear?
- Who benefits most, and whose workload gains little?
- What costs or compatibility constraints could prevent adoption?
- How does this compare with inverse filtering, separate relationship resources, or export for your use case?
- What limitation of per-parent traversal matters to your architecture?
- What is still a design question in the revision you evaluated?
- What evidence supports your optimism about broader identity impact?
- What would you ask the draft author or working group to clarify?
- What can a reader do this week, before standardization or broad client support?
- What should they defer until there is stronger evidence or agreement?

**Read:** [alternatives in draft Section 16](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-16), [the root-level traversal open issue](https://www.ietf.org/archive/id/draft-kushwaha-scim-attr-cursor-pagination-01.html#section-16.9), and the [current Datatracker status](https://datatracker.ietf.org/doc/draft-kushwaha-scim-attr-cursor-pagination/).

**Write:** connect the impact to specific engineering work and identity operations. Distinguish your expectation from an established result. End with a small set of actions, each tied to something the reader can inspect, test, or decide.

**Done when:** the ending answers the opening problem, identifies a limit, and gives a next step without promising inevitable adoption.

**Your section:**

_Write here._

## Write the five takeaways last — 150 words total

Write exactly five bullets, each a complete, self-standing sentence. A takeaway should contain a conclusion or action that the article supports.

Use these questions to draft them yourself:

1. What precise distinction changes how a reader diagnoses an oversized SCIM response?
2. Where should they inspect or change their implementation, and why?
3. What client behavior must they verify to keep membership updates safe?
4. What consistency or compatibility assumption must they resolve before relying on traversal?
5. What should they do next to evaluate this proposal for their own workload?

Check that each sentence teaches something without requiring the reader to follow a teaser. Make the five points distinct. Match their certainty to the body.

**Your five takeaways:**

1. _Write here._
2. _Write here._
3. _Write here._
4. _Write here._
5. _Write here._

## Revise for a colleague, not a specification committee

Read your first draft aloud. Mark sentences you would not naturally say while explaining the problem to another experienced engineer.

- Replace generic introductions with the specific operation or decision.
- Explain SCIM-specific terms when they first affect the reasoning; assume general engineering knowledge.
- Put an important claim next to its evidence and caveat.
- Use first person for your observations, decisions, and judgments.
- Identify an untested design as a recommendation, rather than turning it into a past achievement.
- Prefer a concrete example to repeating the same abstract warning.
- Give alternatives their strongest relevant case.
- Use short lists for parallel checks and procedures; keep explanation in connected paragraphs.
- Remove a paragraph if it adds no evidence, explanation, decision, or useful limitation.
- Keep real outcomes unresolved when the evidence is unresolved.

Ask a peer to identify the strongest technical insight, the least supported claim, and the first action they would take after reading. Their answers should match what you intended.

## Verify sources and examples before submission

The source map here is pinned to revision `-01`, dated August 30, 2026, which Datatracker listed as an individual Internet-Draft when checked on September 9, 2026. Recheck the [proposal record](https://datatracker.ietf.org/doc/draft-kushwaha-scim-attr-cursor-pagination/) before submission. If the revision changes, inspect the changes and update both the article and its section links.

- [ ] I personally read the source behind each substantive protocol claim.
- [ ] I distinguish existing RFC behavior, proposed requirements, non-normative guidance, and my recommendations.
- [ ] I identify the evaluated revision and avoid implying IETF endorsement or guaranteed adoption.
- [ ] I checked example JSON, parameter names, extension identifiers, endpoint scope, and negotiation context.
- [ ] I distinguish examples I ran from illustrative examples.
- [ ] Every result or number has an explanation of where it came from.
- [ ] Claims about client products have version-specific evidence or a checked primary source.
- [ ] The text distinguishes partial, empty, excluded, unauthorized, and complete data where relevant.
- [ ] The article fits 3,000–4,000 words without relying on research notes or this workbook.

## Prepare the InfoQ submission information

Use the author guidelines supplied in this conversation as the submission checklist. Complete these fields yourself; keep them separate from the article.

**Original and unpublished status:** Have you verified whether any version has been publicly accessible? State the facts.

**Proposed title:** What implementation problem and useful insight does the title convey?

**Topic focus and persona:** Which engineering decision does it address, and which InfoQ persona fits best?

**Target reader:** What role, responsibilities, and prior knowledge do you assume?

**Technologies and tools:** List those actually discussed.

**Distinct contribution:** What does your experience or analysis add beyond the draft and existing coverage?

**Real-world basis:** What did you personally implement, investigate, or operate? State explicitly whether you implemented this proposal.

**Cases and use cases:** List each, identifying real cases, experiments, and hypothetical examples.

**Code examples:** List their purpose and whether you tested them.

**Five takeaways:** Copy your final five sentences.

**Authors and co-authors:** Supply everyone's full name.

**Professional biography:** Write one paragraph about relevant experience; include a LinkedIn or other professional profile link.

**Contact information:** Supply each author's email.

**AI use and policy confirmation:** Read the supplied policy and describe the actual creation process. The history includes an initial AI-written article and its conversion into this workbook. Explain what, if anything, you retained from that draft and what assistance you used afterward. Do not describe the entire history as outline-only assistance. Your new article's core writing and technical reasoning must be yours; this workbook does not certify editorial eligibility.

**Image and legal-policy confirmation:** Read the supplied policies and confirm only what you have actually reviewed. For any visuals, record creator, source, license or permission. If none are included, state that.

**Submission timing:** Include a completion timeframe only if submitting an outline instead of a finished article.

**Eligibility:** Check the one-active-proposal rule, quarterly limit, and any consequences of prior declines against your own submission history.

Keep the prospective article unpublished if you intend to seek InfoQ publication. Under the supplied guidelines, publication elsewhere must wait until the four-week exclusivity period has ended and include the required attribution. For a shared draft document, check the requested viewing permissions before submitting. Nothing in this workbook sends a proposal or confirms a policy on your behalf.

## Optional prompts for feedback on your own writing

Use these after you write. They request review, not replacement article sections.

**Practitioner-depth review**

> Review the passage below as a senior engineer evaluating a SCIM integration. Identify unsupported technical claims, missing tradeoffs, and places where an example needs evidence. Ask up to five precise questions. Keep my prose unchanged.

**Argument review**

> Read my draft and identify its main claim, the evidence supporting it, and the strongest counterargument it leaves unanswered. Distinguish problems in reasoning from preferences about style. Suggest where I should add or cut material without writing replacement paragraphs.

**Clarity review**

> Mark sentences that are hard to follow, repetitive, or unnecessarily introductory for experienced engineers. Explain the problem with each and suggest small edits that preserve my meaning. Flag ambiguity for me to resolve; do not invent explanations or experiences.

**Source audit**

> Compare each protocol claim in my draft with the linked primary source and exact revision. Report supported, overstated, contradicted, or unverified claims with section references. Separate proposed behavior from existing standards and from my recommendations. Leave revisions to me.

**Submission review**

> Compare my article and submission notes with the InfoQ guidelines I provide. Identify missing information and statements that require my personal confirmation. Evaluate the disclosed AI use as written; do not infer authorship from writing style or claim that a wording change guarantees acceptance.

Start with the ten experience questions at the top. Once those notes are concrete, draft the storage or client section you know best.
