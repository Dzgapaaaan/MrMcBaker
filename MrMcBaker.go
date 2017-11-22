package main

import (
	Core "MrMcBaker/Core"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token   string
	CfgFile string
	Config  Core.Config
	Parser  Core.Parser
	Logger  Core.Logger
	bot     *discordgo.Session
	err     error
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&CfgFile, "c", "", "Config file")
	flag.Parse()

	Parser, Logger = Config.Init(CfgFile)
	registerCommands(&Parser)
	Parser.LinkLogger(&Logger)
}

func main() {
	bot, err = discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session:\n\t", err)
		return
	}

	bot.AddHandler(onMessage)
	bot.AddHandler(onStatusUpdate)

	err = bot.Open()
	if err != nil {
		fmt.Println("Error opening connection:\n\t", err)
		return
	}

	fmt.Println("Bot is up and running!")
	bot.UpdateStatus(0, Config.Playing)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	bot.Close()
	Config.End(CfgFile, &Parser, &Logger)
}

func onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	//discordgo.Channel().GuildID
	if m.Author.ID == s.State.User.ID {
		Logger.UpdateEntryMsg(m.Author.ID, m)
		return
	}
	s.ChannelMessageSend(m.ChannelID, Parser.Execute(s, m))
	if strings.Contains(m.Content, "🅱") {
		s.MessageReactionAdd(m.ChannelID, m.ID, "🅱")
	}

	Logger.UpdateEntryMsg(m.Author.ID, m)
}

func onStatusUpdate(s *discordgo.Session, p *discordgo.PresenceUpdate) {
	Logger.UpdateEntryPresence(p.Presence.User.ID, p)
}
