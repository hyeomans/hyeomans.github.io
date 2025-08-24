---
title: "ES6 Arrow functions"
date: 2015-09-21T00:00:00-07:00
updated: 2015-09-21T00:00:00-07:00
aliases: ["/es6-arrow-functions"]
author: "Hector Yeomans"
description: "Exploring ES6 arrow functions and their lexical scoping behavior compared to traditional function binding in JavaScript."
tags: ["javascript", "es6", "programming", "arrow-functions"]
series: ["JavaScript & TypeScript Tips"]
series_order: 1
ShowReadingTime: true
ShowBreadCrumbs: true
---

Last night I was reading this post: [ES6 arrow functions, syntax and lexical scoping](http://toddmotto.com/es6-arrow-functions-syntaxes-and-lexical-scoping/) and going through the comments I saw this question:

```
so arrow functions always inherit scope?
```

The answer was by Barney: `always`.

I went to the console and typed:

```
nvm use 4
node
var doSome = () => { console.log(this.x) }
doSome.call({x: 'hello'});
global.x = 'hello';
doSome.call({x: 'good bye'});
```

Could you guess what is going to be printed?

I could replicate this on ES5 without fat arrow function.

```
nvm use 0.10
node
var doSome = function() { console.log(this.x) }.bind(void(0));

doSome.call({x: 'hello'});

global.x = 'hello';
doSome.call({x: 'good bye'});
```

What do you think? What does Babel does?
