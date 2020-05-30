---
title: Convert node package with Typescript and Rollup
date: '2020-05-30T00:00:00Z'
---

I already had [a library](https://github.com/hyeomans/zuora-js) that I wanted to convert to Typescript. I picked Rollup.js to do my build process.

First I installed the following packages:

```bash
> npm i -E -D rollup typescript @rollup/plugin-commonjs @rollup/plugin-node-resolve rollup-plugin-typescript2 rollup-plugin-peer-deps-external
```

I ended up with:

```
* rollup-plugin-typescript2@0.27.1
* rollup@2.11.2
* typescript@3.9.3
* rollup-plugin-peer-deps-external@2.2.2
* @rollup/plugin-node-resolve@8.0.0
* @rollup/plugin-commonjs@12.0.0
```

Then I created a `tsconfig.json`:

```json
{
  "compilerOptions": {
    "outDir": "build",
    "module": "esnext",
    "target": "es5",
    "lib": ["es6", "dom", "es2016", "es2017"],
    "sourceMap": true,
    "allowJs": false,
    "declaration": true,
    "moduleResolution": "node",
    "forceConsistentCasingInFileNames": true,
    "noImplicitReturns": true,
    "noImplicitThis": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "suppressImplicitAnyIndexErrors": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true
  },
  "include": ["src"],
  "exclude": ["node_modules", "build"]
}
```

Also a `rollup.config.js`:

```js
import typescript from 'rollup-plugin-typescript2'
import external from 'rollup-plugin-peer-deps-external'
import commonjs from '@rollup/plugin-commonjs'
import resolve from '@rollup/plugin-node-resolve'

import pkg from './package.json'

export default {
  input: 'src/index.ts',
  output: [
    {
      file: pkg.main,
      format: 'cjs',
      exports: 'named',
      sourcemap: true
    },
    {
      file: pkg.module,
      format: 'es',
      exports: 'named',
      sourcemap: true
    }
  ],
  plugins: [
    external(),
    resolve(),
    typescript({
      rollupCommonJSResolveHack: true,
      exclude: '**/__tests__/**',
      clean: true
    }),
    commonjs({
      include: ['node_modules/**']
    })
  ]
}
```

Modified my `package.json`:

```json
  "main": "build/index.js",
  "module": "build/index.es.js",
  "jsnext:main": "build/index.es.js",
  "files": ["build"],
  "scripts": {
    "build": "rollup -c",
    ...
  }
```

Finally I created a dummy `index.ts` for testing:

```ts
function index() {
  console.log('hello world')
}
```

And ran:

```bash
> npm run build
```

### ESLINT

I had to update eslint packages:

```bash
npm i -E -D @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

Ended up with:

```
* @typescript-eslint/parser@3.0.2
* @typescript-eslint/eslint-plugin@3.0.2
```

And then modified my `.eslintrc.js`:

```js
...
parser: '@typescript-eslint/parser',
  extends: [
    'standard',
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
  ],
  plugins: ['@typescript-eslint'],
  ...
```

You can see the resulting branch here:

https://github.com/hyeomans/zuora-js/tree/ts-rollup

Thanks for reading!