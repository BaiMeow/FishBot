package main

import (
	"flag"
	"fmt"
	auth "github.com/Tnze/go-mc/authenticate"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/mattn/go-colorable"
	"log"
	"time"
)

var (
	c                           *bot.Client
	watch                       chan time.Time
	ip, name, account, password string
	port                        int
	timeout						int
)

func main() {
	log.SetOutput(colorable.NewColorableStdout())
	flag.IntVar(&timeout,"t",45,"自动重新抛竿时间")
	flag.StringVar(&ip, "ip", "localhost", "服务器IP")
	flag.StringVar(&name, "name", "", "游戏ID")
	flag.IntVar(&port, "port", 25565, "端口，默认25565")
	flag.StringVar(&password, "passwd", "", "Mojang账户密码")
	flag.StringVar(&account, "account", "", "Mojang账号")
	flag.Parse()
	log.Println("自动钓鱼机器人启动！")
	log.Println("基于github.com/Tnze/go-mc")
	log.Println("作者: Tnze＆BaiMeow")
	log.Println("-h参数以查看更多用法")
	if account != "" {
		resp, err := auth.Authenticate(account, password)
		if err != nil {
			panic(err)
		}
		if resp.Error != "" {
			log.Println(resp.Error, resp.ErrorMessage)
		}
		log.Println("Auth success")

		c = bot.NewClient()
		c.Auth = bot.Auth{
			Name: resp.SelectedProfile.Name,
			UUID: resp.SelectedProfile.ID,
			AsTk: resp.AccessToken,
		}
	} else {
		c = bot.NewClient()
		c.Auth = bot.Auth{
			Name: name,
			UUID: "",
			AsTk: "",
		}
	}
	//Login
	err := c.JoinServer(ip, port)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Login success")
	go SendMsg()
	//Regist event handlers
	c.Events.GameStart = onGameStart
	c.Events.ChatMsg = onChatMsg
	c.Events.Disconnect = onDisconnect
	c.Events.SoundPlay = onSound

	//JoinGame
	err = c.HandleGame()
	if err != nil {
		log.Fatal(err)
	}
}

func SendMsg() error {
	var send string
	for {
		fmt.Scanln(&send)
		if err := c.Chat(send); err != nil {
			return err
		}
 	}
}

func onGameStart() error {
	log.Println("Game start")

	watch = make(chan time.Time)
	go watchDog()

	return c.UseItem(0)
}

func onSound(name string, category int, x, y, z float64, volume, pitch float32) error {
	if name == "entity.fishing_bobber.splash" {
		if err := c.UseItem(0); err != nil { //retrieve
			return err
		}
		log.Println("gra~")
		time.Sleep(time.Millisecond * 300)
		if err := c.UseItem(0); err != nil { //throw
			return err
		}
		watch <- time.Now()
	}
	return nil
}

func onChatMsg(c chat.Message, pos byte) error {
	log.Println("Chat:", c)
	return nil
}

func onDisconnect(c chat.Message) error {
	log.Println("Disconnect:", c)
	return nil
}

func watchDog() {
	to := time.NewTimer(time.Second * time.Duration(timeout))
	for {
		select {
		case <-watch:
		case <-to.C:
			log.Println("rethrow")
			if err := c.UseItem(0); err != nil {
				panic(err)
			}
		}
		to.Reset(time.Second * time.Duration(timeout))
	}
}
