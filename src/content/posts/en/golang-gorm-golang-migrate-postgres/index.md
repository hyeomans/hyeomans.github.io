---
title: "How to Use GORM and Golang Migrate"
description: "Complete guide on setting up GORM with Golang Migrate to manage PostgreSQL database migrations in Go applications."
pubDate: 2024-06-07T13:00:00-07:00
author: "Hector Yeomans"
tags: ["golang", "gorm", "postgres", "database", "migrations", "programming"]
lang: "en"
draft: false
---

Introduction

This guide demonstrates how to set up and use GORM with Golang Migrate to manage your PostgreSQL database. Follow along to see the code in action and get your environment running.

Prerequisites

1. Go installed on your system.
1. Docker installed and running.

## Project Structure

Here is the final directory structure:

```sh
.
├── .env
├── Makefile
├── cmd
│   └── cli
│       └── main.go
├── docker-compose.yml
├── go.mod
├── go.sum
└── internal
    ├── db
    │   └── migrations
    │       ├── 20240608192206_create_users_table.down.sql
    │       └── 20240608192206_create_users_table.up.sql
    ├── models
    │   └── models.go
    └── repositories
        └── repositories.go

8 directories, 10 files
```

## Step 1: Initialize Your Project

**1. Initialize the Go module:**

```
go mod init github.com/yourusername/yourproject
```

**2. Install dependencies:**

```
go get -u gorm.io/gorm
go get -u gorm.io/driver/postgres
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/file
go get -u github.com/joho/godotenv
```

**3. Set up Docker Compose:**

```yaml
# docker-compose.yml
services:
  gorm-db-post:
    image: postgres:16
    restart: always
    ports:
      - "5433:5432"
    env_file:
      - .env
```

**4.Create the .env file:**

```
# .env
POSTGRES_DB=mydb
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_HOST=localhost
POSTGRES_PORT=5433
```

## Step 2: Define Your Models

```go
// internal/models/models.go
package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"uniqueIndex"`
}

type UserCreateRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
```

## Step 3: Set Up Repositories

```go
// internal/repositories/repositories.go
package repositories

import (
	"context"

	"github.com/yourusername/yourproject/internal/models"
	"gorm.io/gorm"
)

func New(db *gorm.DB) *Repositories {
	return &Repositories{
		User: UserRepository{db: db},
	}
}

type Repositories struct {
	User UserRepository
}

type UserRepository struct {
	db *gorm.DB
}

func (u *UserRepository) Create(ctx context.Context, params models.UserCreateRequest) (*models.User, error) {
	user := models.User{
		Name:  params.Name,
		Email: params.Email,
	}
	if err := u.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
```

## Step 4: Leverage Makefile to write migration files

**1. Install golang-migrate**

If you are on MacOS, you can use `brew`:

```
brew install golang-migrate
```

For other OS's, follow the instructions here: https://github.com/golang-migrate/migrate/tree/master/cmd/migrate

**2.Create a Makefile to run migrate tool**

Use the following `Makefile` to help you generate migrations:

```makefile
include .env
$(eval export $(shell sed -ne 's/ *#.*$$//; /./ s/=.*$$// p' .env))

DOCKER_COMPOSE_FILE := docker-compose.yml

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

## Development

dc-up: ## Start the docker-compose services
	@echo "${GREEN}Starting docker-compose services...${RESET}"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d

dc-down: ## Stop the docker-compose services
	@echo "${GREEN}Stopping docker-compose services...${RESET}"
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down

## Database
db-create-migration: ## TO use pass the name as `make create-migration name=your_migration_name`
	@echo "${GREEN}Creating migration...${RESET}"
	@migrate create -dir internal/db/migrations -ext sql -digits 6 $(name)

## Help:
help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
```

If you run `make help`, you will see:

```
Usage:
  make <target>

Targets:
  Development
    dc-up               Start the docker-compose services
    dc-down             Stop the docker-compose services
  Database
    db-create-migration TO use pass the name as `make create-migration name=your_migration_name`
  Help:
    help                Show this help.
```

**3.Create migrations**

```
make db-create-migration name=create_users_table
```

**4.Fill migration files**

Migration up file:

```sql
-- internal/db/migrations/{date}_create_users_table.up.sql
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

Migration down file:

```sql
-- internal/db/migrations/{date}_create_users_table.down.sql
DROP TABLE IF EXISTS users;
```

## Step 5: Putting it all together

**1. Create the `main.go` file inside `cmd/cli/main.go`:**

```go
// cmd/cli/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hyeomans/gorm-migrate-blogpost/internal/models"
	"github.com/hyeomans/gorm-migrate-blogpost/internal/repositories"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)

	m, err := migrate.New(
		"file://internal/db/migrations",
		dbUrl,
	)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	handleMigrateUp(m)

	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Close the database connection
	repo := repositories.New(db)

	// Create a user
	userReq := models.UserCreateRequest{
		Name:  "John Doe",
		Email: "test1@test.com",
	}
	user, err := repo.User.Create(ctx, userReq)
	if err != nil {
		log.Fatalf("failed to create user: %v", err)
	}
	log.Printf("created user: %v", user)
}

func handleMigrateUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			log.Println("no change")
			return
		}
		log.Fatalf("failed to apply migration: %v", err)
	}
}

```

**2. Start Postgresql**:

In a terminal window, run:

```
make dc-up
```

**3. Run the application**:

```
go run cmd/cli/main.go
```

## Conclusion

By using GORM and Golang Migrate, you can manage your database schema and interact with your PostgreSQL database efficiently. This setup helps maintain versioned migration files and provides better control over database changes.

## Follow-up

- How do you handle migrations in a production environment with Golang Migrate?
- What are some best practices for structuring your models and repositories in a Golang application?
- How can you extend this setup to include more complex relationships and associations between models?
