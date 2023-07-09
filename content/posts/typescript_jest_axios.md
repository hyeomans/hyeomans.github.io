---
title: Typescript, Jest and Axios
date: "2019-11-03T00:00:00Z"
updated: "2019-11-03T00:00:00Z"
aliases: ["/migrate-node-package-typescript-rollup"]
---

I found different posts that tell you how to mock Axios using Jest & Typescript. The only difference in this post is that, when I use Axios, I like to use it as a function rather than calling `axios.get` or `axios.post`.

Imagine you have this Axios request that you want to mock in your tests:

```typescript
//src/index.ts
import axios from "axios";

export interface Post {
  userId: number;
  id: number;
  title: string;
  body: string;
}

const DummyRequest = (id: number): Promise<Post> => {
  return axios({
    method: "GET",
    url: `https://jsonplaceholder.typicode.com/posts/${id}`,
  }).then((response) => {
    return { ...response.data };
  });
};

export default DummyRequest;
```

Install jest and jest-ts and initialize jest-ts

```bash
>npm i -D ts-jest jest
>npx ts-jest config:init
```

This last command will create a `jest.config.js` file:

```js
//jest.config.js
module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
};
```

In your tsconfig.json file, make **sure that your tests are excluded from the compiler**:

```json
//tsconfig.json
...
"exclude": [
      "test/**/*" <--Add this to your exclude array
 ],
...
```

Now we can create a test for our DummyRequest.ts, create this file under `test/index.test.ts`:

```typescript
import axios, { AxiosResponse } from "axios";
import DummyRequest from "../src";
import { mocked } from "ts-jest/dist/util/testing"; //<-- This allows to mock results

jest.mock("axios"); //This is needed to allow jest to modify axios at runtime

it("returns a post", async () => {
  //Arrange
  const axiosResponse: AxiosResponse = {
    data: {
      userId: 1,
      id: 1,
      title:
        "sunt aut facere repellat provident occaecati excepturi optioreprehenderit",
      body: "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto",
    },
    status: 200,
    statusText: "OK",
    config: {},
    headers: {},
  };

  mocked(axios).mockResolvedValue(axiosResponse); //Mocking axios function rather than a method

  //Act
  const result = await DummyRequest(1);

  //Assert
  expect(result).toBe({
    userId: 1,
    id: 1,
    title:
      "sunt aut facere repellat provident occaecati excepturi optioreprehenderit",
    body: "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto",
  });
});
```

Now you can mock the whole Axios function rather than specific methods.
