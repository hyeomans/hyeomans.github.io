---
title: Golang and interfaces misuse
date: "2020-01-18T00:00:00Z"
updated: "2020-01-18T00:00:00Z"
aliases: ["/golang-and-interfaces-misuse"]
---

One of my favorite things about Golang is the concept of interface. It's also one of my grievances every time I see them used as C#/Java interfaces.

It's typical to see a colossal interface defined at a package level file, for example, a package that defines CRUD operations for a User.

```go
package db

import "context"

// User --
type User struct {
 ID int
 Email string
}

// UsersDb pointless interface
type UsersDb interface {
 Get(ctx context.Context, id int) User
 Create(ctx context.Context, user User) User
 Update(ctx context.Context, id int, user User) User
 Delete(ctx context.Context, id int) bool
}
//... then we have the actual struct that "implements" the interface
```

What benefit does this bring to the package? Some say mocking capabilities. From what I've seen to mock that kind of package, you need a generator that generates and updates the mocks.

A better approach is to follow what Golang wiki describes:

> Go interfaces generally belong in the package that uses values of the interface type, not the package that implements those values.

In the case of our `db` package, we defined an interface without knowing if another consumer implements it. But how do we test? And I think that's the wrong question to ask, the right question is, how does my business logic require User information?

Let's assume that we have a simple business requirement: when retrieving the name, the name its always uppercase.

You define a dumb `business` package, so you follow separation of concerns and get something like:

```go
package business

import (
 "context"
 "strings"

 "github.com/hyeomans/interface-misuses/db"
)

// User --
type User struct {
 Name string
}

// Service --
type Service struct {
 //Here you will have logging
 //Counts, etc
 userService db.UserService
}

// NewService --
func NewService(userService db.UserService) Service {
 return Service{
 userService: userService,
 }
}

// GetUser --
func (s Service) GetUser(ctx context.Context, id int) User {
 dbUser := s.userService.Get(ctx, id)
 return User{Name: strings.ToUpper(dbUser.Name)}
}
```

You followed dependency injection and because you defined an `interface` of `db.UserService` you can generate mocks and get testing.

However, there is much boilerplate code to get testing and mocking. There is a more straightforward way to do it, taking advantage of what Golang wiki says:

> Do not define interfaces on the implementor side of an API "for mocking"; instead, design the API so that it can be tested using the public API of the real implementation.
> https://github.com/golang/go/wiki/CodeReviewComments#interfaces

Let's define an interface with only _one_ method in `business.Service`, then instead of injecting the whole `db.UserService`, substitute that with your new interface.

```go
package business

import (
 "context"
 "strings"

 "github.com/hyeomans/interface-misuses/db"
)

// User --
type User struct {
 Name string
}

// UserGetter <---- This interface is defined in the
// consumer side.
// "Do not define interfaces before they are used"
type UserGetter interface {
 Get(ctx context.Context, id int) db.User
}

// Service --
type Service struct {
 userGetter UserGetter // <--- Change this
}

// NewService --
func NewService(userGetter UserGetter) Service {
 return Service{
 userGetter: userGetter, // <----- Change this
 }
}

// GetUser --
func (s Service) GetUser(ctx context.Context, id int) User {
 dbUser := s.userGetter.Get(ctx, id) // You still get a dbUser
 return User{Name: strings.ToUpper(dbUser.Name)}
}
```

Your `main.go` file looks something like:

```go
package main

import (
 "context"
 "fmt"

 "github.com/hyeomans/interface-misuses/business"
 "github.com/hyeomans/interface-misuses/db"
)

func main() {
 ctx := context.Background()
 userService := db.NewUserService()
 service := business.NewService(userService)

 user := service.GetUser(ctx, 1)
 fmt.Println(user.Name)
}
```

## Conclusion

Golang allows the perfect inversion of control, thanks to this pattern. In its essence, it is abstract, but once you start playing with it, it's powerful.

If you need more information about this pattern, there are other great blog posts:

https://github.com/golang/go/wiki/CodeReviewComments#interfaces
https://dave.cheney.net/2016/08/20/solid-go-design
https://twitter.com/davecheney/status/1030790804011245569?lang=en
https://www.ardanlabs.com/blog/2017/07/interface-semantics.html
https://github.com/ardanlabs/gotraining/blob/master/topics/go/design/composition/pollution/example1/example1.go
