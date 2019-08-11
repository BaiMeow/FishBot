package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	_ "github.com/Tnze/go-mc/data/lang/zh-cn"
	"github.com/Tnze/go-mc/realms"
	ygg "github.com/Tnze/go-mc/yggdrasil"
)

const version = "1.14.4"

var (
	c       *bot.Client
	watch   chan time.Time
	timeout int
)

func main() {
	var (
		ip, name, account, password, authserver string
		port                                    int
		realm                                   bool
	)
	flag.IntVar(&timeout, "t", 45, "自动重新抛竿时间")
	flag.StringVar(&ip, "ip", "localhost", "服务器IP")
	flag.StringVar(&name, "name", "", "游戏ID")
	flag.IntVar(&port, "port", 25565, "端口，默认25565")
	flag.StringVar(&password, "passwd", "", "Mojang账户密码")
	flag.StringVar(&account, "account", "", "Mojang账号")
	flag.StringVar(&authserver, "auth", "https://authserver.mojang.com", "验证服务器（外置登陆）")
	flag.BoolVar(&realm, "realms", false, "加入领域服")
	flag.Parse()
	log.Println("自动钓鱼机器人启动！")
	log.Println("基于github.com/Tnze/go-mc")
	log.Println("作者: Tnze＆BaiMeow")
	log.Println("-h参数以查看更多用法")
	if account != "" {
		if authserver != "https://authserver.mojang.com" {
			ygg.AuthURL = fmt.Sprintf("%s/authserver", authserver)
		}
		resp, err := ygg.Authenticate(account, password)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.Println("Auth success")
		if len(resp.AvailableProfiles()) != 1 {
			//多用户选择、登陆
			log.Println("选择角色")
			for i := 0; i < len(resp.AvailableProfiles()); i++ {
				fmt.Printf("[%d]"+resp.AvailableProfiles()[i].Name+"\n", i)
			}
			var no int
			log.Println("请输入角色序号")
			fmt.Scan(&no)
			c = bot.NewClient()
			c.Name = resp.AvailableProfiles()[no].Name
			c.Auth.UUID = resp.AvailableProfiles()[no].ID
			var astk string
			astk, err := resp.Refresh(&resp.AvailableProfiles()[no])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			c.AsTk = astk
		} else {
			//单用户登陆
			c = bot.NewClient()
			c.Auth.UUID, c.Name = resp.SelectedProfile()
			c.AsTk = resp.AccessToken()
		}
	} else {
		//盗版登陆
		c = bot.NewClient()
		c.Auth = bot.Auth{
			Name: name,
			UUID: "",
			AsTk: "",
		}
	}
	//Login
	if authserver != "https://authserver.mojang.com" {
		c.SessionURL = fmt.Sprintf("%s/sessionserver/session/minecraft/join", authserver)
		log.Println("第三方验证")
	} else {
		c.SessionURL = "https://sessionserver.mojang.com/session/minecraft/join"
		log.Println("Mojang验证")
	}
	//判断是否领域服登陆，获取领域服IP
	if realm == true {
		var r *realms.Realms
		r = realms.New(version, c.Name, c.AsTk, c.Auth.UUID)
		servers, err := r.Worlds()
		if err != nil {
			panic(err)
		}
		var i, no = 0, 0
		//list realms
		for _, v := range servers {
			fmt.Printf("[%d]"+v.Name+"\n", i)
			i++
		}
		//agree TOS
		if err := r.TOS(); err != nil {
			panic(err)
		}
		log.Println("请输入领域序号")
		fmt.Scan(&no)
		//GET IP
		ipandport, err := r.Address(servers[no])
		if err != nil {
			panic(err)
		}
		s := strings.Split(ipandport, ":")
		ip = s[0]
		port, _ = strconv.Atoi(s[1])
	}
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
		//fmt.Scanln(&send)
		Reader := bufio.NewReader(os.Stdin)
		send, _ = Reader.ReadString('\n')
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
	log.Println("Chat:", c.String())
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
