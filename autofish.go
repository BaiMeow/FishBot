package main

import (
	"flag"
	"fmt"
	bot "github.com/Tnze/gomcbot"
	"github.com/Tnze/gomcbot/authenticate"
	"log"
	"math"
	"time"
)

func main() {
	var online, offline bool
	flag.BoolVar(&online, "online-mode", false, "是否启用正版验证")
	flag.BoolVar(&offline, "offline-mode", false, "是否不启用正版验证")
	var account, passwd string
	flag.StringVar(&account, "account", "", "Mojang账号")
	flag.StringVar(&passwd, "passwd", "", "Mojang账号的密码")
	var name, uuid string
	flag.StringVar(&name, "name", "Steve", "游戏内名称")
	flag.StringVar(&uuid, "uuid", "", "玩家识别码")
	var server string
	var port int
	flag.StringVar(&server, "ip", "localhost", "加入的服务器IP")
	flag.IntVar(&port, "port", 25565, "服务器端口")

	flag.Parse()

	log.Println("自动钓鱼机器人启动！")
	log.Println("基于github.com/Tnze/gomcbot")
	log.Println("作者: Tnze")
	log.Println("-h参数以查看更多用法")
	var auth bot.Auth

	if !online && !offline {
		for {
			log.Println("是否启动正版验证:(true/false)")
			n, err := fmt.Scanf("%t\n", &online)
			if n == 1 && err == nil {
				break
			}
			log.Printf("请输入%q或%q\n", "true", "false")
		}
	}

	if online { //online mode
		if account == "" {
			log.Println("请输入账号:")
			fmt.Scanf("%s\n", &account)
		}
		if passwd == "" {
			log.Println("请输入密码:")
			fmt.Scanf("%s\n", &passwd)
		}
		resp, err := authenticate.Authenticate(account, passwd)
		if err != nil {
			log.Fatalf("登录失败: %v\n", err)
		}
		auth = resp.ToAuth()
		log.Println("登录成功")
	} else { //offline mode
		auth = bot.Auth{
			Name: name,
			UUID: uuid,
		}
	}
	g, err := auth.JoinServer(server, port)
	if err != nil {
		log.Fatalf("加入服务器%s:%d失败: %v\n", server, port, err)
	}
	log.Println("成功加入服务器")

	//Handle game
	events := g.GetEvents()
	go g.HandleGame()

	//处理事件
	for e := range events {
		switch e {
		case bot.PlayerSpawnEvent:
			log.Println("Player is spawn! 开始钓鱼")
			go fish(g)
		}
	}
}

func fish(g *bot.Game) {
	g.SetSoundCallBack(func(s int32, c int32, x, y, z float64, v, p float32) {
		if s == 184 { // 184 是钓到鱼的声音！
			p := g.GetPlayer()
			x0, y0, z0 := p.GetPosition()
			if length(x-x0, y-y0, z-z0) < 5 {
				log.Println("钓到物品")

				g.UseItem(true) //收竿
				time.Sleep(time.Millisecond * 500)
				g.UseItem(true) //下一次甩竿
			} else {
				log.Printf("远处的钓鱼声(%.2f, %.2f, %.2f), %g\n", x-x0, y-y0, z-z0, length(x-x0, y-y0, z-z0))
			}
		}
	})
	g.UseItem(true) //甩竿
}

func length(x, y, z float64) float64 {
	return math.Abs(x) +
		math.Abs(y) +
		math.Abs(z)
}
