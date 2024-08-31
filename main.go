package main

import (
	"fmt"
	"log"
	"net/http"

	"strings"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/gowon-irc/go-gowon"
	"github.com/imroc/req/v3"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	APIKey string `short:"k" long:"api-key" env:"GOWON_RA_API_KEY" required:"true" description:"retroachievements api key"`
	KVPath string `short:"K" long:"kv-path" env:"GOWON_RA_KV_PATH" default:"kv.db" description:"path to kv db"`
}

const (
	moduleName = "retroachievements"
	moduleHelp = "get players last achievements from retroachievements"
)

func setUser(kv *bolt.DB, nick, user []byte) error {
	err := kv.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("retroachievements"))
		return b.Put([]byte(nick), []byte(user))
	})
	return err
}

func getUser(kv *bolt.DB, nick []byte) (user []byte, err error) {
	err = kv.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("retroachievements"))
		v := b.Get([]byte(nick))
		user = v
		return nil
	})
	return user, err
}

func parseArgs(msg string) (command, user string) {
	fields := strings.Fields(msg)

	if len(fields) >= 1 {
		command = fields[0]
	}

	if len(fields) >= 2 {
		user = fields[1]
	}

	return command, user
}

func setUserHandler(kv *bolt.DB, nick, user string) (string, error) {
	if user == "" {
		return "Error: username needed", nil
	}

	err := setUser(kv, []byte(nick), []byte(user))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("set %s's user to %s", nick, user), nil
}

type commandFunc func(*req.Client, string) (string, error)

func CommandHandler(client *req.Client, kv *bolt.DB, nick, user string, f commandFunc) (string, error) {
	if user != "" {
		return f(client, user)
	}

	savedUser, err := getUser(kv, []byte(nick))
	if err != nil {
		return "", err
	}

	if len(savedUser) == 0 {
		return "Error: username needed", nil
	}

	return f(client, string(savedUser))
}

func raHandler(client *req.Client, kv *bolt.DB, m *gowon.Message) (string, error) {

	command, user := parseArgs(m.Args)

	switch command {
	case "s", "set":
		return setUserHandler(kv, m.Nick, user)
	case "a", "achievement":
		return CommandHandler(client, kv, m.Nick, user, raNewestAchievement)
	case "l", "last":
		return CommandHandler(client, kv, m.Nick, user, raLastGames)
	case "c", "current":
		return CommandHandler(client, kv, m.Nick, user, raCurrentStatus)
	case "p", "points":
		return CommandHandler(client, kv, m.Nick, user, raPoints)
	case "w", "awards":
		return CommandHandler(client, kv, m.Nick, user, raAwards)
	case "g", "game":
		return CommandHandler(client, kv, m.Nick, user, raGameProgress)
	}

	return "one of [s]et, [a]chievement, [l]ast, [c]urrent, [p]oints, a[w]ards or [g]ame must be passed as a command", nil
}

func main() {
	log.Printf("%s starting\n", moduleName)

	opts := Options{}
	if _, err := flags.Parse(&opts); err != nil {
		log.Fatal(err)
	}

	kv, err := bolt.Open(opts.KVPath, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer kv.Close()

	err = kv.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("retroachievements"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	httpClient := req.C().
		SetCommonQueryParam("y", opts.APIKey)

	r := gin.Default()
	r.POST("/message", func(c *gin.Context) {
		var m gowon.Message

		if err := c.BindJSON(&m); err != nil {
			log.Println("Error: unable to bind message to json", err)
			return
		}

		out, err := raHandler(httpClient, kv, &m)
		if err != nil {
			log.Println(err)
			m.Msg = "{red}Error when looking up retroachievements data{clear}"
			c.IndentedJSON(http.StatusInternalServerError, &m)
		}

		m.Msg = out
		c.IndentedJSON(http.StatusOK, &m)
	})

	r.GET("/help", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, &gowon.Message{
			Module: moduleName,
			Msg:    moduleHelp,
		})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
