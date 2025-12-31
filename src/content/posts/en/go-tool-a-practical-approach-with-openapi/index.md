---
title: "Managing Go Tools the Right Way: From tools.go to go tool"
description: "Go 1.24 introduces a better way to manage project tools. Let's explore the evolution from manual installs to tools.go to the new go tool directive using OpenAPI code generation as a practical example."
pubDate: 2025-12-31T02:51:37+00:00
author: "Hector Yeomans"
tags: ["golang", "tooling"]
lang: "en"
draft: false
heroImage: "./hero.jpeg"
heroAlt: "Hero image for Go Tool: A practical approach with OpenAPI"
---

Before Go `1.24`, managing project tools like code generators, linters, and formatters was awkward. You had two options:

1. **Manual Installation**: Require developers to know to run `make install-deps` or `make deps` or something similar before they can work on the project.
2. **The `tools.go` pattern**: A file with blank imports to track tool dependencies.

You can read more about the `tools.go` approach here: https://www.jvt.me/posts/2022/06/15/go-tools-dependency-management/

The `tools.go` approach had its own issues:

* **Performance hit**: `go run` invocations were not cached, so repeated calls were slow. Usually this is fine because you are not running tools constantly. At least that was the case for me in my previous projects.
* **Dependency bloat**: Tool dependencies polluted your `go.mod`, and consumers of your module would see them as indirect dependencies.

## How `go tool` Works

Rather than explaining the theory, let's build a practical project. We'll do it the old way first and then migrate to `go tool`.

For this example I will show three ways you can use the [oapi-codegen generator](https://github.com/oapi-codegen/oapi-codegen):

1. Installing the binary globally (you can uninstall at the end)
2. Using the `tools.go` pattern
3. Using the new `go tool` directive

## Setting Up the Project

First create a new project and initialize it:

```sh
mkdir go-tools-demo && cd go-tools-demo && go mod init github.com/hyeomans/tooldemo
```

## Approach 1: Installing the Binary

Let's create a simple Makefile to install dependencies. Create a `Makefile` and add this:

```make
GOBIN := $(shell go env GOPATH)/bin

.PHONY: install-deps remove-deps generate

install-deps:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1
	mv $(GOBIN)/oapi-codegen $(GOBIN)/oapi-codegenv2

remove-deps:
	rm -f $(GOBIN)/oapi-codegenv2

generate:
	oapi-codegenv2 --config misc/oapi-config.yml misc/openapi.yml
```

Why rename the binary? Sometimes you need multiple versions of the same tool, or you want to be explicit about which version you're using. This is a simple way to avoid conflicts.

Run `make install-deps` and you will see something like this:

```
$ make install-deps 
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1
go: downloading github.com/oapi-codegen/oapi-codegen/v2 v2.5.1
go: downloading gopkg.in/yaml.v2 v2.4.0
go: downloading github.com/speakeasy-api/openapi-overlay v0.10.2
go: downloading github.com/getkin/kin-openapi v0.133.0
...
mv ~/go/bin/oapi-codegen ~/go/bin/oapi-codegenv2
```

Now we need an OpenAPI spec for the generator to work with. Create `misc/openapi.yml`:

```yml
openapi: 3.0.3
info:
  title: Tasks API
  version: 1.0.0
  description: A simple task management API

servers:
  - url: http://localhost:8080

paths:
  /tasks:
    get:
      operationId: listTasks
      summary: List all tasks
      responses:
        '200':
          description: A list of tasks
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Task'
    post:
      operationId: createTask
      summary: Create a new task
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTaskRequest'
      responses:
        '201':
          description: Task created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Task'

  /tasks/{id}:
    get:
      operationId: getTask
      summary: Get a task by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: The task
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Task'
        '404':
          description: Task not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
    delete:
      operationId: deleteTask
      summary: Delete a task
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '204':
          description: Task deleted
        '404':
          description: Task not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

components:
  schemas:
    Task:
      type: object
      required:
        - id
        - title
        - completed
        - createdAt
      properties:
        id:
          type: string
          format: uuid
        title:
          type: string
          example: Buy groceries
        description:
          type: string
          example: Milk, eggs, bread
        completed:
          type: boolean
          default: false
        createdAt:
          type: string
          format: date-time

    CreateTaskRequest:
      type: object
      required:
        - title
      properties:
        title:
          type: string
          example: Buy groceries
        description:
          type: string
          example: Milk, eggs, bread

    Error:
      type: object
      required:
        - message
      properties:
        message:
          type: string
          example: Task not found
```

You also need a config file for the OpenAPI generator. Create `misc/oapi-config.yml`:

```yml
package: api
output: cmd/service/api/gen.go
generate:
  models: true
  chi-server: true
```

Create the output directory and run the generator:

```sh
mkdir -p cmd/service/api
make generate
```

This works, but it requires everyone on your team to run `make install-deps` before they can generate code. Let's look at better aproaches.

## Approach 2: The tools.go Pattern

First, remove the binary we installed:

```sh
make remove-deps
```

If you try to run `make generate` now, it will fail because the binary is gone.

The `tools.go` pattern lets us track tool dependencies in our module. Create `misc/tools.go`:

```go
//go:build tools
// +build tools

package main

import (
	_ "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen"
)
```

The build constraint ensures this file is never actually compiled into your binary. It only exists to tell Go about the dependency.

Run `go mod tidy` and check your `go.mod`. You will see:

```
require (
	github.com/oapi-codegen/oapi-codegen/v2 v2.5.1
)
```

Now instead of calling the binary directly, we use `go run`. Create `cmd/service/api/generate.go`:

```go
package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../../misc/oapi-config.yml ../../../misc/openapi.yml
```

Modify the `misc/oapi-config.yml` file:

```yml
package: api
output: gen.go
generate:
  models: true
  chi-server: true
```

Update your Makefile:

```make
generate:
	go generate ./...
```

Now run it:

```sh
rm cmd/service/api/gen.go
go generate ./...
```

This is better because new developers only need to clone the repo and run `go generate`. No manual installation required. However, it has the downsides I mentioned earlier: no caching and dependency bloat.

## Approach 3: The New go tool Directive

Go 1.24 introduces a cleaner solution. First, remove the `tools.go` file:

```sh
rm misc/tools.go
```

Now add the tool using the new `-tool` flag:

```sh
go get -tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1
```

Check your `go.mod` and you will see a new `tool` directive at the bottom:

```
tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
```

Update your `generate.go` file to use `go tool` instead of `go run`:

```go
package api

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=../../../misc/oapi-config.yml ../../../misc/openapi.yml
```

Notice we replaced `go run` with `go tool`. This is shorter and the tool invocations are now cached.

Test it:

```sh
rm cmd/service/api/gen.go
go generate ./...
```

## Which Approach Should You Use?

For new projects using Go 1.24 or later, I recommend `go tool`. It gives you:

* Cached invocations (faster subsequent runs)
* Clear separation between runtime and tool dependencies
* No need for the `tools.go` workaround

For projects that need to support older Go versions, stick with the `tools.go` pattern. It works and is well understood.

I still find value in keeping the Makefile around as an entry point for common tasks:

```make
GOBIN := $(shell go env GOPATH)/bin

.PHONY: install-deps remove-deps generate

install-deps:
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1
	mv $(GOBIN)/oapi-codegen $(GOBIN)/oapi-codegenv2

remove-deps:
	rm -f $(GOBIN)/oapi-codegenv2

generate:
	go generate ./...
```

The `install-deps` target is still useful for CI environments or when you want a globally availble binary for quick testing.

## Conclusion

The `go tool` directive is a welcome addition to Go's toolchain. It solves real problems that the community has been working around for years. If you're starting a new project with Go 1.24, give it a try.

You can find the complete example code on [GitHub](https://github.com/hyeomans/tooldemo).
