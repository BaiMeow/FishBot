package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"time"

	"github.com/MscBaiMeow/FishBot/net"
	"github.com/google/uuid"

	"github.com/MscBaiMeow/FishBot/float"
	l "github.com/MscBaiMeow/FishBot/mclogin"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	_ "github.com/Tnze/go-mc/data/lang/zh-cn"
	ygg "github.com/Tnze/go-mc/yggdrasil"
	"github.com/mattn/go-colorable"
)

var (
	c       *bot.Client
	watch   chan time.Time
	timeout int
	resp    *ygg.Access
	auth    l.Player
	version = "1.16.1"
	waiting bool
)

func main() {
	var (
		ip, name, account, authserver string
		port                          int
		realm, removemode             bool
		uuid                          string
	)
	log.SetOutput(colorable.NewColorableStdout())
	flag.IntVar(&timeout, "t", 45, "自动重新抛竿时间")
	flag.StringVar(&ip, "ip", "", "服务器IP")
	flag.StringVar(&name, "name", "", "游戏ID")
	flag.IntVar(&port, "port", 25565, "端口，默认25565")
	flag.StringVar(&account, "account", "", "Mojang账号")
	flag.StringVar(&authserver, "auth", "https://authserver.mojang.com", "验证服务器（外置登陆）")
	flag.BoolVar(&realm, "realms", false, "加入领域服")
	flag.StringVar(&uuid, "uuid", "", "直接根据uuid读取配置文件进入游戏")
	flag.BoolVar(&removemode, "remove", false, "删除配置模式")
	flag.Parse()
	log.Println("自动钓鱼机器人启动！")
	log.Println("机器人版本：1.6.0")
	log.Printf("游戏版本：%s", version)
	log.Println("基于github.com/Tnze/go-mc")
	log.Println("作者: Tnze＆BaiMeow")
	log.Println("-h参数以查看更多用法")
	if ip == "" && (!realm && !removemode) {
		log.Fatal("没有指定服务器IP，请使用-ip指定，或使用-realms进入领域")
	}
	net.CheckSRV(&ip, &port)
	switch {
	//验证登陆
	case account != "" && name == "" && uuid == "" && removemode == false:
		l.Authlogin(&account, &authserver, &auth)
		//盗版登陆
	case account == "" && name != "" && uuid == "" && removemode == false:
		auth.Name = name
		auth.Authmode = "Offline"
		//读取配置直接选择账户
	case account == "" && name == "" && uuid == "" && removemode == false:
		l.LoadConfiglogin(&auth)
		//没有输入登陆信息则读取配置询问账户
	case account == "" && name == "" && uuid != "" && removemode == false:
		l.Directlogin(uuid, &auth)
	case account == "" && name == "" && uuid == "" && removemode == true:
		l.Configrm()
	default:
		log.Fatal("格式错误：请使用以下四种格式之一\n验证登陆：AutoFish -account xxx \n盗版登陆：AutoFish -name xxx\n根据uuid读取配置登陆：AutoFish -uuid xxx\n读取配置询问登陆：AutoFish\n删除模式：AutoFish -remove")
	}

	c = bot.NewClient()
	c.SettingMapRead(false)
	c.Name, c.Auth.UUID, c.AsTk = auth.Name, auth.UUID, auth.Tokens.AccessToken
	l.AddtoConfig(&auth)
	//判断是否领域服登陆，整一个领域ip
	if realm && auth.Authmode != "Offline" {
		if err := l.Checkrealms(&ip, &port, c, &version, &auth); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}
	for {
		//Join Game
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
		c.Events.SpawnObj = onSpawnObj
		c.Events.EntityRelativeMove = onEntityRelativeMove

		//JoinGame
		err = c.HandleGame()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Reconnect in 5s")
		time.Sleep(5 * time.Second)
	}
}
func sendMsg() error {
	var send []byte
	for {
		Reader := bufio.NewReader(os.Stdin)
		send, _, _ = Reader.ReadLine()
		if err := c.Chat(string(send)); err != nil {
			return err
		}
	}
}

func onGameStart() error {
	log.Println("Game start")
	watch = make(chan time.Time)
	go watchDog()
	time.Sleep(time.Second)
	err := c.Chat("Login with MscFishBot.")
	if err != nil {
		return err
	}
	return c.UseItem(0)
}

func onChatMsg(msg chat.Message, pos byte, sender uuid.UUID) error {
	log.Println("Chat:", msg.String())
	return nil
}

func onDisconnect(msg chat.Message) error {
	log.Println("Disconnect:", msg)
	return nil
}

func watchDog() {
	to := time.NewTimer(time.Second * time.Duration(timeout))
	for {
		select {
		case <-watch:
		case <-to.C:
			log.Println("rethrow")
			Held := c.HeldItem
			if err := c.SelectItem((Held + 1) % 9); err != nil {
				log.Printf("fail to select item:%s", err.Error())
			}
			time.Sleep(time.Millisecond * 500)
			if err := c.SelectItem(Held); err != nil {
				panic(err)
			}
			time.Sleep(time.Millisecond * 500)
			if err := c.UseItem(0); err != nil {
				log.Printf("fail to use item:%s", err.Error())
			}
		}
		to.Reset(time.Second * time.Duration(timeout))
	}
}

func onSpawnObj(EID int, UUID [16]byte, Type int, x, y, z float64, Pitch, Yaw float32, Data int, VelocityX, VelocitY, VelocitZ int16) error {
	//	fmt.Println(Type)
	if Type == 106 {
		if Data == c.EntityID {
			//	log.Println("Spawn your Float")
			float.Set(EID, x, y, z)
			//	} else {
			//	log.Println("Other's Float")
		}
	}
	return nil
}

func onEntityRelativeMove(EID, DeltaX, DeltaY, DeltaZ int) error {
	if float.IsMine(EID) {
		if float.IsFish(DeltaX, DeltaY, DeltaZ) {
			getFish()
		}
	}
	return nil
}

func getFish() error {
	if err := c.UseItem(0); err != nil { //retrieve
		return err
	}
	log.Println("gra~")
	time.Sleep(time.Millisecond * time.Duration(timeout*8))
	if err := c.UseItem(0); err != nil { //throw
		return err
	}
	watch <- time.Now()
	return nil
}
