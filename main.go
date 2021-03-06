package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	commands =  []*discordgo.ApplicationCommand{
		{
			Name: "makecategory",
			Description: "Command which designates a role as a category",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "category",
					Description: "Role to become a category",
					Required: true,
				},
			},
		},
		{
			Name: "setcategory",
			Description: "Command which assigns roles to a category",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "category",
					Description: "Category to assign role to",
					Required: true,
				},
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "role",
					Description: "Role to assign to category",
					Required: true,
				},
			},
		},
		{
			Name: "removecategory",
			Description: "Command which removes a role from the category list",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "category",
					Description: "Category to remove",
					Required: true,
				},
			},
		},
		{
			Name: "updatecategory",
			Description: "Command which changes the category a role is assigned to",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "role",
					Description: "Role to change",
					Required: true,
				},
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "category",
					Description: "Category to change to",
					Required: true,
				},
			},
		},
		{
			Name: "unsetcategory",
			Description: "Command which removes a role's category",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type: discordgo.ApplicationCommandOptionRole,
					Name: "role",
					Description: "Role to change",
					Required: true,
				},
			},
		},
		{
			Name: "listall",
			Description: "Lists every category and its roles",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
		"makecategory": func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
			margs := []interface{}{
				i.Data.Options[0].RoleValue(nil, "").ID,
			}
			var msgformat string
			mr, err := checkManageRoles(i.Member, i.ChannelID, s)
			if err != nil || mr == false {
				if err != nil {
					msgformat = err.Error()
				} else {
					msgformat = "User does not have manage roles permission!"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			}
			err = addCategory(i.Data.Options[0].RoleValue(nil, "").ID, i.GuildID, db)
			if err != nil {
				msgformat = err.Error()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			} else {
				msgformat = `role <@&%s> is now a category!`
				// disable mentions by passing a zero'd allowmentions
				var allowedMentions discordgo.MessageAllowedMentions
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						AllowedMentions: &allowedMentions,
						Content: fmt.Sprintf(
							msgformat,
							margs...,
						),
					},
				})
			}
		},
		"setcategory": func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
			margs := []interface{}{
				i.Data.Options[0].RoleValue(nil, "").ID,
				i.Data.Options[1].RoleValue(nil, "").ID,
			}
			var msgformat string
			mr, err := checkManageRoles(i.Member, i.ChannelID, s)
			if err != nil || mr == false {
				if err != nil {
					msgformat = err.Error()
				} else {
					msgformat = "User does not have manage roles permission!"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			}
			err = setCategory(i.Data.Options[0].RoleValue(nil, "").ID, i.Data.Options[1].RoleValue(nil, "").ID, i.GuildID, db)
			if err != nil {
				msgformat = err.Error()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
			} else {
				msgformat = `category <@&%s> now contains role <@&%s>`
				// disable mentions by passing a zero'd allowmentions
				var allowedMentions discordgo.MessageAllowedMentions
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						AllowedMentions: &allowedMentions,
						Content: fmt.Sprintf(
							msgformat,
							margs...,
						),
					},
				})
			}
		},
		"removecategory": func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
			margs := []interface{}{
				i.Data.Options[0].RoleValue(nil, "").ID,
			}
			var msgformat string
			mr, err := checkManageRoles(i.Member, i.ChannelID, s)
			if err != nil || mr == false {
				if err != nil {
					msgformat = err.Error()
				} else {
					msgformat = "User does not have manage roles permission!"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			}
			err = removeCategory(i.Data.Options[0].RoleValue(nil, "").ID, i.GuildID, db)
			if err != nil {
				msgformat = err.Error()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
			} else {
				msgformat = `<@&%s> is no longer a category!`
				// disable mentions by passing a zero'd allowmentions
				var allowedMentions discordgo.MessageAllowedMentions
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						AllowedMentions: &allowedMentions,
						Content: fmt.Sprintf(
							msgformat,
							margs...,
						),
					},
				})
			}
		},
		"updatecategory": func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
			margs := []interface{}{
				i.Data.Options[0].RoleValue(nil, "").ID,
				i.Data.Options[1].RoleValue(nil, "").ID,
			}
			var msgformat string
			mr, err := checkManageRoles(i.Member, i.ChannelID, s)
			if err != nil || mr == false {
				if err != nil {
					msgformat = err.Error()
				} else {
					msgformat = "User does not have manage roles permission!"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			}
			err = updateCategory(i.Data.Options[1].RoleValue(nil, "").ID, i.Data.Options[0].RoleValue(nil, "").ID, i.GuildID, db)
			if err != nil {
				msgformat = err.Error()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
			} else {
				msgformat = `<@&%s> is now part of <@&%s>!`
				// disable mentions by passing a zero'd allowmentions
				var allowedMentions discordgo.MessageAllowedMentions
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						AllowedMentions: &allowedMentions,
						Content: fmt.Sprintf(
							msgformat,
							margs...,
						),
					},
				})
			}
		},
		"unsetcategory": func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
			margs := []interface{}{
				i.Data.Options[0].RoleValue(nil, "").ID,
			}
			var msgformat string
			mr, err := checkManageRoles(i.Member, i.ChannelID, s)
			if err != nil || mr == false {
				if err != nil {
					msgformat = err.Error()
				} else {
					msgformat = "User does not have manage roles permission!"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			}
			err = unsetCategory(i.Data.Options[0].RoleValue(nil, "").ID, i.GuildID, db)
			if err != nil {
				msgformat = err.Error()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
			} else {
				msgformat = `<@&%s> is no longer part of a category!`
				// disable mentions by passing a zero'd allowmentions
				var allowedMentions discordgo.MessageAllowedMentions
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						AllowedMentions: &allowedMentions,
						Content: fmt.Sprintf(
							msgformat,
							margs...,
						),
					},
				})
			}
		},
		"listall": func(s *discordgo.Session, i *discordgo.InteractionCreate, db *mongo.Client) {
			var msgformat string
			mr, err := checkManageRoles(i.Member, i.ChannelID, s)
			if err != nil || mr == false {
				if err != nil {
					msgformat = err.Error()
				} else {
					msgformat = "User does not have manage roles permission!"
				}
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
				return
			}
			roles, err := listRoles(i.GuildID, db)
			if err != nil {
				msgformat = err.Error()
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						Content: fmt.Sprintf(
							msgformat,
						),
					},
				})
			} else {
				var embed discordgo.MessageEmbed
				embed.Color = 0xCC00CC
				embed.Title = "Categories"
				embed.Description = ""
				for _, n := range roles {
					embed.Description += "<@&" + n.Category + ">\n"
					for _, k := range n.Roles {
						embed.Description += "<@&" + k + "> "
					}
					embed.Description += "\n"
				}
				var embeds []*discordgo.MessageEmbed
				embeds = append(embeds, &embed)
				// disable mentions by passing a zero'd allowmentions
				var allowedMentions discordgo.MessageAllowedMentions
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionApplicationCommandResponseData{
						AllowedMentions: &allowedMentions,
						Embeds: embeds,
					},
				})
			}
		},
	}
)

const (
	address = "localhost:50051"
	port    = ":50051"
)

type guildCategories struct {
	Guild string
	Categories []categoryHolder
}

type guildRoles struct {
	Guild string
	Roles []roleHolder
}

// i am so horrible at naming things
type catRoles struct {
	Category string
	Roles []string
}

type roleHolder struct {
	Role string
	Category string
}

type categoryHolder struct {
	Role string
}

func checkManageRoles(member *discordgo.Member, channelID string, s *discordgo.Session) (perms bool, err error) {
	channel, err := s.State.Channel(channelID)
	if err != nil {
		return
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return
	}

	if member.User.ID == guild.OwnerID {
		perms = true
		return
	}

	for _, role := range guild.Roles {
		for _, roleID := range member.Roles {
			if role.ID == roleID {
				if role.Permissions&discordgo.PermissionManageRoles == discordgo.PermissionManageRoles {
					perms = true
					return
				}
			}
		}
	}

	return
}

func listRoles(gid string, db *mongo.Client) (ret []catRoles, err error) {
	rolesCollection := db.Database("test").Collection("roles")
	categoriesCollection := db.Database("test").Collection("categories")

	var gCats guildCategories
	var gRoles guildRoles

	filter := bson.D{{"guild", gid}}
	err = categoriesCollection.FindOne(context.Background(), filter).Decode(&gCats)
	if err == mongo.ErrNoDocuments {
		err = errors.New("No categories to list! Register a category with /makecategory")
		return
	}
	if err != nil {
		return
	}

	err = rolesCollection.FindOne(context.Background(), filter).Decode(&gRoles)
	if err == mongo.ErrNoDocuments {
		for _, n := range gCats.Categories {
			var tmp catRoles
			tmp.Category = n.Role
			ret = append(ret, tmp)
		}
		return
	}
	if err != nil {
		return
	}


	for _, n := range gRoles.Roles {
		found := false
		for i, k := range ret {
			if k.Category == n.Category {
				ret[i].Roles = append(ret[i].Roles, n.Role)
				found = true
				break
			}
		}
		if found == false {
			var tmp catRoles
			tmp.Category = n.Category
			tmp.Roles = append(tmp.Roles, n.Role)
			ret = append(ret, tmp)
		}
	}
	for _, n := range gCats.Categories {
		found := false
		for _, k := range ret {
			if k.Category == n.Role {
				found = true
				break
			}
		}
		if found == false {
			var tmp catRoles
			tmp.Category = n.Role
			ret = append(ret, tmp)
		}
	}
	return
}

func unsetCategory(role string, gid string, db *mongo.Client) error {
	collection := db.Database("test").Collection("roles")

	filter :=  bson.D{{"guild", gid}}
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$pull", bson.D{{"roles", bson.D{{"role", role}}}}}})
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Did not unset any categories!")
	}
	return nil
}

func updateCategory(cat string, role string, gid string, db *mongo.Client) error {
	collection := db.Database("test").Collection("roles")

	filter := bson.D{{"guild", gid}}
	var gCat guildCategories
	// check if new category is a category
	err := db.Database("test").Collection("categories").FindOne(context.Background(), filter).Decode(&gCat)
	if err == mongo.ErrNoDocuments {
		return errors.New("No categories registered! Register a category with /makecategory")
	}
	if err != nil {
		return err
	}
	found := false
	for _, k := range gCat.Categories {
		if k.Role == cat {
			found = true
		}
	}
	if found == false {
		return errors.New("Category is not a category!")
	}

	arrayFilter := options.Update().SetArrayFilters(options.ArrayFilters{Filters: []interface{}{bson.M{"elem.role": role}}})
	res, err := collection.UpdateOne(context.Background(),
		filter,
		bson.D{{"$set", bson.D{{"roles.$[elem].category", cat}}}},
		arrayFilter)
	if err == mongo.ErrNoDocuments {
		return errors.New("No roles set! Set a role to a category with /setcategory")
	}
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Did not update any roles! Is the role registered to a category?")
	}
	return nil
}

func removeCategory(cat string, gid string, db *mongo.Client) error {
	collection := db.Database("test").Collection("categories")

	filter := bson.D{{"guild", gid}}
	res, err := collection.UpdateOne(context.Background(), filter, bson.D{{"$pull", bson.D{{"categories", bson.D{{"role", cat}}}}}})
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return errors.New("Did not delete any categories!")
	}

	collection = db.Database("test").Collection("roles")
	res, err = collection.UpdateOne(context.Background(), filter, bson.D{{"$pull", bson.D{{"roles", bson.D{{"category", cat}}}}}})
	if err != nil {
		return err
	}

	return nil
}

func setCategory(cat string, role string, gid string, db *mongo.Client) error {
	collection := db.Database("test").Collection("roles")
	filter := bson.D{{"guild", gid}}

	var gCats guildCategories

	err := db.Database("test").Collection("categories").FindOne(context.Background(), filter).Decode(&gCats)
	if err == mongo.ErrNoDocuments {
		return errors.New("No categories registered! Register a category with /makecategory")
	}
	if err != nil {
		return err
	}

	// check if this role is a category
	// and that this category is a category
	found := false
	for _, k := range gCats.Categories {
		if k.Role == role {
			return errors.New("Role is a category!")
		}
		if k.Role == cat {
			found = true
			break
		}
	}
	if found == false {
		return errors.New("Category is not a category!")
	}

	var gRoles guildRoles

	// check if this role already has a category
	err = collection.FindOne(context.Background(), filter).Decode(&gRoles)
	if err == mongo.ErrNoDocuments {
		_, errnew := collection.InsertOne(context.Background(), bson.M{"guild": gid, "roles": bson.A{bson.D{{"role", role}, {"category", cat}}}})
		return errnew
	}
	if err != nil {
		return err
	}

	for _, k := range gRoles.Roles {
		if k.Role == role {
			return errors.New("Role already belongs to a category!")
		}
	}

	var ins roleHolder
	ins.Role = role
	ins.Category = cat
	_, err = collection.UpdateOne(context.Background(), filter, bson.D{{"$push", bson.D{{"roles", ins}}}})
	if err != nil {
		return err
	}

	return nil
}

func addCategory(cat string, gid string, db *mongo.Client) error {
	collection := db.Database("test").Collection("categories")

	var guild guildCategories

	// check if this role is already a category
	filter := bson.D{{"guild", gid}}
	err := collection.FindOne(context.Background(), filter).Decode(&guild)
	if err == mongo.ErrNoDocuments {
		_, errnew := collection.InsertOne(context.Background(), bson.M{"guild": gid, "categories": bson.A{bson.D{{"role", cat}}}})
		return errnew
	}
	if err != nil {
		return err
	}

	for _, n := range guild.Categories {
		if n.Role == cat {
			return errors.New("Role is already a category!")
		}
	}

	var ins categoryHolder
	ins.Role = cat
	_, err = collection.UpdateOne(context.Background(), filter, bson.D{{"$push", bson.D{{"categories", ins}}}})
	if err != nil {
		return err
	}

	return nil
}

// The error handling here is bad. I don't know how to improve it.
func guildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate, db *mongo.Client) {
	// Find all of the roles a member has which are stored in the db
	collection := db.Database("test").Collection("roles")

	var gRoles guildRoles

	filter := bson.D{{"guild", m.GuildID}}
	err := collection.FindOne(context.Background(), filter).Decode(&gRoles)
	if err != nil {
		fmt.Println(err)
		return
	}

	var roles []roleHolder
	for _, n := range m.Roles {
		for _, k := range gRoles.Roles {
			if n == k.Role {
				roles = append(roles, k)
			}
		}
	}

	// Find all the roles a member has that are considered categories
	collection = db.Database("test").Collection("categories")

	var gCat guildCategories
	err = collection.FindOne(context.Background(), filter).Decode(&gCat)
	if err != nil {
		fmt.Println(err)
		return
	}

	var categories []categoryHolder
	for _, n := range m.Roles {
		for _, k := range gCat.Categories {
			if n == k.Role {
				categories = append(categories, k)
			}
		}
	}

	// if a user has a role that requires a category
	// but does not have that category, add that category to the user
	for _, role := range roles {
		found := false
		for _, category := range categories {
			if (role.Category == category.Role) {
				found = true
				break
			}
		}

		if found == false {
			err = s.GuildMemberRoleAdd(m.GuildID, m.User.ID, role.Category)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

	// if a user has a category
	// check that it has roles that require it
	// remove the category if the user does not
	for _, category := range categories {
		found := false
		for _, role := range roles {
			if (role.Category == category.Role) {
				found = true
				break
			}
		}

		if found == false {
			err = s.GuildMemberRoleRemove(m.GuildID, m.User.ID, category.Role)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// registers new commands as soon as a guild is joined
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, event.Guild.ID, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	collection := mongoClient.Database("test").Collection("token")

	var result struct {
		Token string
	}
	err = collection.FindOne(context.Background(), bson.D{}).Decode(&result)
	if err != nil {
		log.Fatal(err)
	}

	if result.Token == "" {
		fmt.Println("No token found, please insert one into DB!")
		return
	}

	discord, err := discordgo.New("Bot " + result.Token)
	if err != nil {
		log.Fatalf("error creating Discord session, %v", err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
		guildMemberUpdate(s, m, mongoClient)
	})
	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i, mongoClient)
		}
	})
	discord.AddHandler(guildCreate)

	discord.Identify.Intents = discordgo.IntentsGuildMembers | discordgo.IntentsGuilds

	err = discord.Open()
	if err != nil {
		log.Fatalf("error opening connection, %v", err)
		return
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	discord.Close()

	err = mongoClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}
