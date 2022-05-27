package main

import (
	"flag"
	"github.com/bwmarrin/discordgo"
	"github.com/darenliang/dsfs/fuse"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var (
	token   string
	guildID string
	mount   string
	compact bool
	options fuseOpts
	// We need to jankily expose the db and dsfs for messageCreate
	db   *DB
	dsfs *Dsfs
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&guildID, "s", "", "Guild ID")
	flag.StringVar(&mount, "m", "", "Mount point for Linux/macOS")
	flag.BoolVar(&compact, "c", false, "Compact transactions")
	flag.Var(&options, "o", "FUSE options")
	flag.Parse()
}

func main() {
	// If DEBUG is set, expose pprof
	if len(os.Getenv("DEBUG")) != 0 {
		go func() {
			log.Println("pprof running on port 8000")
			_ = http.ListenAndServe(":8000", nil)
		}()
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	db, err = setupDB(dg, guildID)
	if err != nil {
		log.Println(err)
		return
	}

	dsfs = NewDsfs(dg, db)
	host := fuse.NewFileSystemHost(dsfs)
	host.SetCapReaddirPlus(true)
	host.Mount(mount, options.Args())
}
