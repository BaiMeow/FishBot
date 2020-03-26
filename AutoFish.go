package main

import (
	"bufio"
	"flag"
	"log"
	"math"
	"os"
	"time"

	"github.com/MscBaiMeow/FishBot/net"

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
	float   floats
	resp    *ygg.Access
	auth    l.Player
	version = "1.15.2"
)

type floats struct {
	ID int
	x  float64
	y  float64
	z  float64
}

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
	log.Println("机器人版本：1.5.0")
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
	c.Events.SoundPlay = onSound
	c.Events.SpawnObj = onSpawnObj
	c.Events.EntityRelativeMove = onEntityRelativeMove
	//c.Events.WindowsItemChange = onWindowsItemChange
	//JoinGame
	err = c.HandleGame()
	if err != nil {
		log.Fatal(err)
	}
}
func sendMsg() error {
	var send []byte
	for {
		//fmt.Scanln(&send)
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

func onSound(name string, category int, x, y, z float64, volume, pitch float32) error {
	if name == "entity.fishing_bobber.splash" {
		if distance(x, y, z) < 5 {
			if err := c.UseItem(0); err != nil { //retrieve
				return err
			}
			log.Println("gra~")
			time.Sleep(time.Millisecond * time.Duration(timeout*8))
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
	x0 := float.x - x
	y0 := float.y - y
	z0 := float.z - z
	return math.Sqrt(x0*x0 + y0*y0 + z0*z0)
}

func onChatMsg(c chat.Message, pos byte) error {
	log.Println("Chat:", c.String())
	return nil
}

func onDisconnect(c chat.Message) error {
	log.Println("Disconnect:", c)
	return nil
}

func watchDog() {
	to := time.NewTimer(time.Second * time.Duration(timeout))
	shouldHeld := c.HeldItem
	nextSlot := (shouldHeld + 1) % 9
	for {
		select {
		case <-watch:
		case <-to.C:
			log.Println("rethrow")
			if err := c.SelectItem(nextSlot); err != nil {
				log.Printf("fail to select item:%s", err.Error())
			}
			time.Sleep(time.Millisecond * 300)
			if err := c.SelectItem(shouldHeld); err != nil {
				log.Printf("fail to select item:%s", err.Error())
			}
			time.Sleep(time.Millisecond * 300)
			if err := c.UseItem(0); err != nil {
				log.Printf("fail to use item:%s", err.Error())
			}
		}
		to.Reset(time.Second * time.Duration(timeout))
	}
}

func onSpawnObj(EID int, UUID [16]byte, Type int, x, y, z float64, Pitch, Yaw float32, Data int, VelocityX, VelocitY, VelocitZ int16) error {
	if Type == 102 {
		if Data == c.EntityID {
			//log.Println("Spawn your Float")
			float = floats{EID, x, y, z}
			//} else {
			//	log.Println("Other's Float")
		}
	}
	return nil
}

func onEntityRelativeMove(EID, DeltaX, DeltaY, DeltaZ int) error {
	if EID == float.ID {
		float.x += float64(DeltaX) / 4096
		float.y += float64(DeltaY) / 4096
		float.z += float64(DeltaZ) / 4096
	}
	return nil
}

//func onWindowsItemChange(id byte, slotID int, slot entity.Slot) error {
//	if id == 0 {
//		fmt.Println(slot.ItemID)
//	}
//	return nil
//}
