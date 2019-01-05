---
title: Resilience with Disyuntor - Circuit Breaker
date: '2017-05-27'
---

Resilience means the capacity to recover quickly from difficulties. Circuit breaker pattern is a good practice for resilience.


When working with distributed systems, you want resilience. If you're working with "micro-services," 
you probably have faced with the problem of a service going down. When X service goes down, and Y and Z depend on X, 
every internal exception could potentially start taking other services down. 

If you don't work with micro-services, you might still have an integration with a payment provider 
(PayPal, Stripe, Google Play, etc.). What happens when any of those providers goes down? 
Imagine a request comes to your internal service, then your service makes a request to Stripe, 
then Stripe takes 30 seconds to tell you there was something wrong. How many requests have 
queued up in 30 seconds in your service?

Circuit Breaker is a pattern that can help you with this problem. The Circuit Breaker pattern 
became famous after Release It! book. To be honest the first time I heard about this 
patterns was in 2011 on Dependency Injection in .NET.

> "Circuit Breaker is a stability pattern because it adds robustness to an application by failing 
fast, instead of hanging and consuming resources while it hangs. This is a good example of a nonfunctional 
requirement and a true CROSS-CUTTING CONCERN, because it has little to do with the feature implemented with the out-of-process call."

>Excerpt From: Mark Seemann. "Dependency Injection in .NET." 2011

_

> "Residential fuses have gone the way of the rotary dial telephone. Now, circuit breakers 
protect overeager gadget hounds from burning their houses down. The principle is the same: 
detect excess usage, fail first, and open the circuit. More abstractly, the circuit breaker exists 
to allow one subsystem (an electrical circuit) to fail (excessive current draw, possibly from a 
short-circuit) without destroying the entire system (the house). Furthermore, once the danger has 
passed, the circuit breaker can be reset to restore full function to the system."

>Excerpt From: Michael T. Nygard. "Release It!" 2007

There are three states on a Circuit Breaker implementation:

* Open
* Half-Open
* Closed

The closed state represents a healthy system. Going back to the Stripe example, 
the closed state means requests come and go without the known existence of a Circuit breaker.

Circuit breaker takes passive action when that HTTP call fails. On every failure, 
the circuit breaker is listening for failures. The Circuit Breaker opens when the 
threshold of failures, or rate of failures, is met.

Once the circuit breaker is open, every HTTP call will fail immediately, bypassing 
the real call to Stripe. After a pre-defined period, the Circuit Breaker tries a 
real call to Stripe, leaving the Circuit Breaker on a half-open state. 

On a half-open state, if the request to Stripe succeeds, the Circuit Breaker returns 
to a closed state, if it fails it returns to an open state.

Usually, an open state call is a custom exception. When using a Circuit Breaker 
implementation, make sure you log and monitor this kind of exceptions.

## Disyuntor
[Disyuntor](https://github.com/auth0/disyuntor) is an implementation of Circuit Breaker in Node.js by [Auth0](https://auth0.com/). This npm package lets you wrap a critical function in a Circuit Breaker pattern.

In this tutorial, you will create two services. One of them will be flaky for a deterministic 
period. The other will issue requests.

After that, we will add Disyuntor and wrap the call in a Circuit Breaker pattern. You will see the three states in action.

### Pre-requisites
This is a Node.js tutorial, but also I will use yarn to install packages. Whenever you see `yarn add --exact {package}` can
be replace with `npm install --exact {package}`. Also I'm doing this in macOS Sierra, so this is a *nix OS. I will try my 
best to make it cross-platform.

Let's create a new project, open up your console and type:

```
$> mkdir disyuntor-example && cd $_
$disyuntor-example> yarn init -y #or npm init -y
```

We will use Express.js to mock out our two services:

```
$disyuntor-example> yarn add --exact express
```

Create the flaky server first: 

```js
// flaky.js
const app = require('express')();

app.get('/:id', (req, res, next) => {
  var param = req.params.id;
  if (param === "0") {
    blockFor(5);
    res.sendStatus(503);
  } else {
    res.status(200).send('I am ok now.');
  }
});

function blockFor(seconds) {
  var waitTill = new Date(new Date().getTime() + seconds * 1000);
  while(waitTill > new Date()){}
}

app.listen(3000, () => console.log('Flaky app is listening on port 3000'));
```

For the sake of this tutorial you will use a simple parameter to control if the server is flaky or not. You will notice the function `blockFor(seconds)`, this was added to simulate a service that takes time to return.

Before creating the other service, you need to add npm packages to create http request to the flaky service. Also you will add a helper package to run both services from a single command:

```bash
$disyuntor-example> yarn add --exact got bluebird concurrently
```

Now let's create our consumer service:

```js
// consumer.js
const app     = require('express')();
const got     = require('got');

app.get('/:id', (req, res, next) => {
  return got(`http://localhost:3000/${req.params.id}`)
    .then(() => {
      res.sendStatus(200);
    })
    .catch(() => {
      res.sendStatus(503);
    });
});

app.listen(4000, () => console.log('Consumer service is listening on port 4000'));
```

As a final step add a `start` under `scripts` into your `package.json`:

```json
...
"scripts": {
  "start": "concurrently \"node flaky\" \"node consumer\" "
},
...
```

Going back to your terminal window, you can type:

```bash
$disyuntor-example> npm start
> concurrently "node flaky" "node consumer"

[0] Flaky app is listening on port 3000
[1] Consumer service is listening on port 4000
```

Both services are running, now in a different terminal window you can make a Curl request:

```bash
$disyuntor-example> curl -I http://localhost:4000/1
HTTP/1.1 200 OK
X-Powered-By: Express
Content-Type: text/plain; charset=utf-8
Content-Length: 2
ETag: W/"2-nOO9QiTIwXgNtWtBJezz8kv3SLc"
Connection: keep-alive

```

Now try to do a request that you know it will return 503:

```bash
$disyuntor-example> curl -I http://localhost:4000/0
HTTP/1.1 503 Service Unavailable
X-Powered-By: Express
Content-Type: text/plain; charset=utf-8
Content-Length: 19
ETag: W/"13-/70LdyMNgL+PAJa+Q/RtnRF82z8"
Date: Sun, 28 May 2017 14:26:29 GMT
Connection: keep-alive
```

Now imagine this is a production setup. Your flaky service is an internal service that has gone down. Your public service start swamping with requests your internal service. After a while your public service becomes unresponsive. 

You can even reproduce that scenario with this command:

```bash
$disyuntor-example> curl -I http://localhost:4000/0 && curl -I http://localhost:4000/0 && curl -I http://localhost:4000/0
```

You will notice that takes more than 10 seconds to finish all the requests.

You can stop both services by pressing Ctrl+C.

Let's add Disyuntor to circuit break this requests:

```bash
$disyuntor-example> yarn add --exact disyuntor
```

Modify your consumer service:

```js
const app       = require('express')();
const got       = require('got');
const disyuntor = require('disyuntor');

const safeGot = disyuntor.promise(got, {
  name: 'got.request',
  timeout: '10s',
  cooldown: '5s',
  maxFailures: 1,
  onTrip: (err, failures, cooldown) => console.log(`got.request triped because it failed ${failures} times. Last error was ${err.message}! There will be no more attempts for ${cooldown}ms`)
});

app.get('/:id', (req, res, next) => {
  return safeGot(`http://localhost:3000/${req.params.id}`)
    .then(() => {
      res.sendStatus(200);
    })
    .catch(() => {
      res.sendStatus(500);
    });
});

app.listen(4000, () => console.log('Consumer service is listening on port 4000'));
```

Start both services again with `npm start`. Start sending requests with curl:

```bash
$disyuntor-example> curl -I http://localhost:4000/0
```

You will see this message where you started your services:

```bash
[1] got.request triped because it failed 1 times. Last error was Response code 503 (Ser
vice Unavailable)! There will be no more attempts for 5000ms
```

Let's try this:

```bash
$disyuntor-example> curl -I http://localhost:4000/0 && curl -I http://localhost:4000/0 && curl -I http://localhost:4000/0
```

You will notice that the first request takes the expected 5 seconds but subsequent request fail immediatedly. This is the Circuit Breaker pattern in action. 

After the first failure, the circuit becomes open. Then after the _cooldown_ period the circuit becomes half-open. If we issue another request after 5 seconds, you will see that it tries again to contact the flaky service.

## Conclusion

When working with multiple external services -- either a Db or http service -- a good resilience practice is to add a circuit breaker.

Disyuntor is a good circuit breaker library, it lacks of some features, but it gets the work done. 
