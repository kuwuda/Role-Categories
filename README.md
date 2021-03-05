# Role Categories
A discord bot for automatically sorting roles into categories

Uses Golang, MongoDB and DiscordGo

Based around Discord's slash-commands, which are beta right now. May change significantly due to that.

## Usage
This bot works by assigning Discord roles to other roles classified as categories.

Requires Manage Roles permissions for both the bot and the user trying to execute the command.

Usage example:

![Example Screenshot](/screenshots/ex.png?raw=true "Example Screenshot")

In this screenshot, "+ Custom Roles +", "+ House Roles +", and "+ Club Roles +" are defined as categories.
The roles under them belong to those categories.
If I were to remove all of the roles under "+ House Roles +" (Director, Benefactor, Housemate) the bot would automatically remove the role "+ House Roles +" because I don't have any roles which are a part of that category.
Likewise, if I readded any of those roles, the bot would automatically add the role "+ House Roles +"

### Creating a category
Using the command /makecategory \[role\] will turn a given role into a category

### Assigning a role to a category
Using the command /setcategory \[category\] \[role\] will assign a role to a category

The rest of the commands should be obvious.

If you want an easy way to make categories that look like the ones in the screenshot, check out 
[my role generator](https://kuwuda.github.io/Discord-Role-Category-Generator/rolecategorygenerator.html)

## Installing
Install requirements: MongoDB, Golang, DiscordGo, MongoDb's go driver. Check the individual instructions on those!

Create a Discord bot and get its token.

Add that token to the db
```
mongo
db.token.insertOne({token: "your token here"})
```

Get this repo:
```
go get github.com/kuwuda/role-categories
```

Compile the bot & run it!
```
go build main.go
./main
```
