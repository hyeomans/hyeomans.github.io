---
title: "Knex.js transaction + pg-boss: keep API writes and jobs consistent"
description: "A practical knex.js transaction + pgboss pattern that keeps database writes and pg-boss jobs consistent in one PostgreSQL transaction."
pubDate: 2026-02-20T17:00:10+00:00
author: "Hector Yeomans"
tags: ["nodejs", "express", "knexjs", "pgboss", "pg-boss", "postgresql", "transactions"]
lang: "en"
draft: false
heroImage: "./hero.jpg"
heroAlt: "Hero image for Knex.js transaction + pg-boss tutorial"
---

Most [Express](https://expressjs.com/) APIs do two things when handling a request. They write to the database. Then they enqueue a background job.

The problem is that these two operations are separate. Your database write can succeed while your queue send fails. Or the other way around. You end up with inconsistent state. A widget exists in your database but no job was queued. Or a job was queued for a widget that never got created.

This causes real problems downstream. Workers process events for entities that do not exist. Or entities exist with no events to trigger notifications or side effects.

If you are looking for a reliable `knex.js transaction pgboss` pattern, this is the one I use: run both operations inside the same [PostgreSQL transaction](https://www.postgresql.org/docs/current/tutorial-transactions.html).

## The problem with the naive approach

Here is what most code looks like.

```js
async function createWidget(data) {
  const widget = await db("widgets").insert(data).returning("*");
  await pgBoss.send("widget.events", { type: "created", widgetId: widget.id });
  return widget;
}
```

This looks fine. It works most of the time. But there are two failure modes.

First, the database insert succeeds. Then something goes wrong with [pg-boss](https://www.npmjs.com/package/pg-boss). Maybe the connection drops. Maybe the queue schema has an issue. The widget exists in your database but no event was published. Downstream systems never hear about it.

Second, pg-boss succeeds. Then the database transaction fails to commit. Maybe a constraint violation happens after the insert. Now you have a queued job pointing to a widget that does not exist. Your worker will try to process a phantom entity.

Both cases leave you with inconsistent state. You could try to handle this with retries or compensating actions. But that gets messy fast. You end up with more code to maintain and more edge cases to debug.

## What we actually need

We want atomicity. Either both operations succeed or neither does. If the widget insert fails, no job should be queued. If the job enqueue fails, the widget insert should roll back.

PostgreSQL transactions give us this guarantee. But pg-boss normally uses its own database connection. It does not know about your [Knex.js transaction](https://knexjs.org/guide/transactions.html).

The good news is that pg-boss supports a custom database interface through the [`db` option](https://github.com/timgit/pg-boss). You can pass a `db` option to `send()` that tells it how to execute SQL. If you give it a wrapper that uses your knex transaction, pg-boss will run its inserts on the same connection. Everything becomes part of one transaction.

## How to make it work

The key is the `db` option in pg-boss. It expects an object with an `executeSql` method. This method receives a SQL string and an array of values. It should return the query results.

Here is the wrapper I use.

```js
_createTxDbWrapper(trx) {
  return {
    executeSql: async (text, values) => {
      const connection = await trx.client.acquireConnection();
      try {
        const result = await trx.client.query(connection, {
          sql: text,
          bindings: values ?? [],
        });
        return result.response;
      } finally {
        await trx.client.releaseConnection(connection);
      }
    },
  };
}
```

This wrapper takes a knex transaction and exposes the `executeSql` interface that pg-boss expects. It borrows the underlying connection from the transaction. It runs the query on that connection. Then it releases the connection back to the transaction.

Notice that we acquire and release the connection manually. We do not use `trx.raw()` because knex might try to open a new connection. We want to reuse the exact same connection that the transaction holds open.

With this wrapper in place, we can generate the options for pg-boss.

```js
_pgBossOptions(trx) {
  const eventId = randomUUID();
  return {
    eventId,
    options: {
      db: this._createTxDbWrapper(trx),
      singletonKey: eventId,
    },
  };
}
```

I generate a UUID for the event. I use that same UUID as the `singletonKey`. This gives us deduplication for free. If somehow the same event gets queued twice, pg-boss will only keep one copy.

Now the repository method can use these options inside a transaction.

```js
async create(ctx, data) {
  return await this.db.transaction(async (trx) => {
    const [widget] = await trx("widgets")
      .insert({ name: data.name, status: data.status || "active" })
      .returning("*");

    const { eventId, options } = this._pgBossOptions(trx);
    await this.pgBoss.send(
      "widget.events",
      { eventId, eventType: "widget.v1.created", entityId: widget.id },
      options
    );

    return { widget, eventId };
  });
}
```

The widget insert and the pg-boss send happen inside `db.transaction()`. If either one fails, knex rolls back the entire transaction. The widget row gets removed. The queued job gets removed. Nothing is left in a half done state.

## What you get from this

The main benefit is consistency. You can trust that if a widget exists in your database, an event for it was queued. If a job exists in your queue, the entity it references definitely exists.

This simplifies your worker code. Workers do not need to handle missing entities as special cases. They can assume the entity exists because the transaction guaranteed it.

It also simplifies error handling in your API. You do not need compensating transactions or cleanup logic. If something fails, the database rolls everything back. You catch the error and return a 500. Done.

Testing becomes easier too. Your tests can verify that both the database write and the queue insert happened in the same transaction. You can simulate failures in pg-boss and confirm that the database write was rolled back.

```js
it("should rollback widget insert when enqueue fails", async () => {
  const failingPgBoss = {
    send: vi.fn().mockRejectedValue(new Error("enqueue failed")),
  };
  const failingRepo = new WidgetsRepository(db, failingPgBoss);

  await expect(
    failingRepo.create(ctx, { name: "Rollback Widget" }),
  ).rejects.toThrow();

  const rows = await db("widgets").where({ name: "Rollback Widget" });
  expect(rows).toHaveLength(0);
});
```

The test confirms that when pg-boss fails, no widget is left behind in the database.

## When to use this pattern

This approach works well when your application uses PostgreSQL for both your primary data and your job queue. Pg-boss stores jobs in PostgreSQL tables. That is what makes this transactional approach possible.

If you use a separate queue system like Redis or RabbitMQ, you cannot do this. Those systems do not participate in PostgreSQL transactions. You would need a different strategy, like the [transactional outbox pattern](https://microservices.io/patterns/data/transactional-outbox.html) or saga orchestration.

But if you are already using PostgreSQL and pg-boss, this pattern is worth adopting. It keeps your code simple. It gives you strong consistency guarantees. It removes a whole class of bugs related to partial failures.

## A few things to keep in mind

The transaction stays open until pg-boss finishes its insert. Normally this is fast. But if your pg-boss schema is on a different database or has latency issues, your transaction will hold locks longer. Watch your query times.

Also remember that pg-boss needs to be initialized with the same database credentials. It needs access to its schema and tables. The transaction wrapper just tells it which connection to use for a specific send operation.

Finally, this pattern works for single operations. If you need to send multiple jobs in the same transaction, you can call `send()` multiple times with the same transaction wrapper. Each call will reuse the same connection.

## Knex.js transaction + pgboss checklist

- Start one `db.transaction()` in Knex.js.
- Insert your data with the same `trx` object.
- Call `pgBoss.send()` with a custom `db.executeSql` wrapper that uses that transaction connection.
- Commit once both the write and the enqueue succeed; otherwise let the transaction roll back everything.

## Summary

Running pg-boss sends inside knex transactions gives you atomicity between database writes and job queue writes. You build a small wrapper that exposes `executeSql` on top of a knex transaction. You pass that wrapper to pg-boss via the `db` option. Now both operations participate in the same transaction.

The result is simpler code and fewer edge cases. You stop thinking about what happens when one succeeds and the other fails. The database handles it for you.

## Further reading

- [Knex.js transactions guide](https://knexjs.org/guide/transactions.html)
- [pg-boss on npm](https://www.npmjs.com/package/pg-boss)
- [pg-boss GitHub repository](https://github.com/timgit/pg-boss)
- [PostgreSQL transactions tutorial](https://www.postgresql.org/docs/current/tutorial-transactions.html)
