package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"

	auth "github.com/Tnze/go-mc/authenticate"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/mattn/go-colorable"
)

var (
	c       *bot.Client
	watch   chan time.Time
	timeout int
)

func main() {
	var (
		ip, name, account, password, authserver string
		port                                    int
	)
	log.SetOutput(colorable.NewColorableStdout())
	flag.IntVar(&timeout, "t", 45, "自动重新抛竿时间")
	flag.StringVar(&ip, "ip", "localhost", "服务器IP")
	flag.StringVar(&name, "name", "", "游戏ID")
	flag.IntVar(&port, "port", 25565, "端口，默认25565")
	flag.StringVar(&password, "passwd", "", "Mojang账户密码")
	flag.StringVar(&account, "account", "", "Mojang账号")
	flag.StringVar(&authserver, "auth", "", "验证服务器（外置登陆）")
	flag.Parse()
	log.Println("自动钓鱼机器人启动！")
	log.Println("基于github.com/Tnze/go-mc")
	log.Println("作者: Tnze＆BaiMeow")
	log.Println("-h参数以查看更多用法")
	if account != "" {
		resp, err := auth.AuthWithThirdPartyServer(account, password, authserver)
		if err != nil {
			panic(err)
		}
		if resp.Error != "" {
			log.Println(resp.Error, resp.ErrorMessage)
		}
		log.Println("Auth success")
		if len(resp.AvailableProfiles) != 1 {
			//	log.Println("选择角色")
			//	for i := 0; i < len(resp.AvailableProfiles); i++ {
			//		fmt.Printf("[%[1]d]"+resp.AvailableProfiles[i].Name+"\n", i)
			//	}
			//	var no int
			//	fmt.Scanf("%d", no)
			//	resp.SelectedProfile.Name = resp.AvailableProfiles[no].Name
			//	resp.SelectedProfile.ID = resp.AvailableProfiles[no].ID
			//	resp.SelectedProfile.Legacy = resp.AvailableProfiles[no].Legacy
			//	err := c.Refresh()
			panic("暂不支持多角色")
		}
		c = bot.NewClient()
		c.Auth = bot.Auth{
			Name:   resp.SelectedProfile.Name,
			UUID:   resp.SelectedProfile.ID,
			AsTk:   resp.AccessToken,
			Server: authserver,
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
	go sendMsg()
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

func sendMsg() error {
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
		if distance(x, y, z) < 4 {
			if err := c.UseItem(0); err != nil { //retrieve
				return err
			}
			log.Println("gra~")
			time.Sleep(time.Millisecond * 300)
			if err := c.UseItem(0); err != nil { //throw
				return err
			}
			watch <- time.Now()
		} else {
			log.Println("远处的钓鱼声")
		}
	}
	return nil
}

func distance(x, y, z float64) float64 {
	x0 := math.Abs(c.X - x)
	y0 := math.Abs(c.Y - y)
	z0 := math.Abs(c.Z - z)
	return math.Sqrt(x0*x0 + y0*y0 + z0*z0)
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

//
