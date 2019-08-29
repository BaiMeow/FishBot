package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	_ "github.com/Tnze/go-mc/data/lang/zh-cn"
	"github.com/Tnze/go-mc/realms"
	ygg "github.com/Tnze/go-mc/yggdrasil"
	"github.com/mattn/go-colorable"
)

const version = "1.14.4"

var (
	c       *bot.Client
	watch   chan time.Time
	timeout int
	float   floats
	auth    player
	resp    *ygg.Access
	conf    config
)

type floats struct {
	ID int
	x  float64
	y  float64
	z  float64
}

type config struct {
	Players []player `json:"players"`
}

type player struct {
	Name       string `json:"name"`
	UUID       string `json:"UUID"`
	AsTk       string `json:"AsTk"`
	Account    string `json:"account"`
	Authserver string `json:"authserver"`
	Authmode   string `json:"authmode"`
}

//0 mojang
//1 three
//2 offline

func main() {
	var (
		ip, name, account, authserver string
		port                          int
		realm                         bool
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
	flag.Parse()
	log.Println("自动钓鱼机器人启动！")
	log.Println("机器人版本：1.4-Pre3")
	log.Printf("游戏版本：%s", version)
	log.Println("基于github.com/Tnze/go-mc")
	log.Println("作者: Tnze＆BaiMeow")
	log.Println("-h参数以查看更多用法")
	if ip == "" {
		log.Fatal("没有指定服务器IP，请使用-ip指定")
	}
	switch {
	//验证登陆
	case account != "" && name == "" && uuid == "":
		authlogin(&account, &authserver)
		//盗版登陆
	case account == "" && name != "" && uuid == "":
		auth.Name = name
		auth.Authmode = "Offline"
		//读取配置直接选择账户
	case account == "" && name == "" && uuid == "":
		loadconfiglogin()
		//没有输入登陆信息则读取配置询问账户
	case account == "" && name == "" && uuid != "":
		directlogin(uuid)
	default:
		log.Fatal("格式错误：请使用以下四种格式之一\n验证登陆：AutoFish -account xxx \n盗版登陆：AutoFish -name xxx\n根据uuid读取配置登陆：AutoFish -uuid xxx\n读取配置询问登陆：AutoFish")
	}

	c = bot.NewClient()
	c.Name, c.Auth.UUID, c.AsTk = auth.Name, auth.UUID, auth.AsTk
	updateconfig()
	//判断是否领域服登陆，整一个领域ip
	if realm == true && auth.Authmode != "Offline" {
		if err := checkrealms(&ip, &port); err != nil {
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
		if distance(x, y, z) < 0.25 {
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

func onSpawnObj(EID int, UUID [16]byte, Type int, x, y, z float64, Pitch, Yaw float32, Data int, VelocityX, VelocitY, VelocitZ int16) error {
	if Type == 101 {
		if Data == c.EntityID {
			//log.Println("Spawn your Float")
			float = floats{EID, x, y, z}
			//} else {
			//	log.Println("Other's Float")
		}
	}
	return nil
}

func onEntityRelativeMove(EID int, DeltaX, DeltaY, DeltaZ int16) error {
	if EID == float.ID {
		float.x += float64(DeltaX) / 4096
		float.y += float64(DeltaY) / 4096
		float.z += float64(DeltaZ) / 4096
	}
	return nil
}

func checkrealms(ip *string, port *int) error {
	var r *realms.Realms
	r = realms.New(version, c.Name, c.AsTk, c.Auth.UUID)
	servers, err := r.Worlds()
	if err != nil {
		return err
	}
	var i, no = 0, 0
	//list realms
	for _, v := range servers {
		fmt.Printf("[%d]"+v.Name+"\n", i)
		i++
	}
	//agree TOS
	if err := r.TOS(); err != nil {
		return err
	}
	log.Println("请输入领域序号")
	fmt.Scan(&no)
	//GET IP
	ipandport, err := r.Address(servers[no])
	if err != nil {
		return err
	}
	s := strings.Split(ipandport, ":")
	*ip = s[0]
	*port, _ = strconv.Atoi(s[1])
	return nil
}

//func onWindowsItemChange(id byte, slotID int, slot entity.Slot) error {
//	if id == 0 {
//		fmt.Println(slot.ItemID)
//	}
//	return nil
//}

func authlogin(account, authserver *string) {
	if *authserver != "https://authserver.mojang.com" {
		ygg.AuthURL = fmt.Sprintf("%s/authserver", *authserver)
		bot.SessionURL = fmt.Sprintf("%s/sessionserver/session/minecraft/join", *authserver)
		log.Println("第三方验证")
		auth.Authmode = "ThreeAuth"
	} else {
		auth.Authmode = "MojangAuth"
	}
	auth.Authserver = *authserver
	auth.Account = *account
	password := ""
	prompt := &survey.Password{
		Message: "Please type your password",
	}
	survey.AskOne(prompt, &password)
	resp, err := ygg.Authenticate(*account, string(password))
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
		auth.Name = resp.AvailableProfiles()[no].Name
		auth.UUID = resp.AvailableProfiles()[no].ID
		if err = resp.Refresh(&resp.AvailableProfiles()[no]); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		auth.AsTk = resp.AccessToken()
		return
	}
	//单用户登陆
	auth.UUID, auth.Name = resp.SelectedProfile()
	auth.AsTk = resp.AccessToken()
	return
}

func loadconf() config {
	data, err := ioutil.ReadFile("./conf.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Fatal("配置文件不存在，请使用-account参数添加", err)
			os.Exit(1)
		}
		log.Fatal(err)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &conf)
	if err != nil {
		log.Fatal(err)
	}
	return conf
}

func loadconfiglogin() {
	conf = loadconf()
	log.Println("load config success")
	//让玩家选一个
	var (
		selected    string
		i           = 0
		preSelected []string
		selectedNo  int
	)
	for _, v := range conf.Players {
		preSelected = append(preSelected, fmt.Sprintf("[%d]", i)+v.Name+"\t"+v.Authmode+"\t"+v.Authserver)
		i++
	}
	prompt := &survey.Select{
		Message: "Choose a profile:",
		Options: preSelected,
	}
	survey.AskOne(prompt, &selected)
	selectedNo, _ = strconv.Atoi(selected[1:2])
	auth = conf.Players[selectedNo]
	//检查登陆类型
	switch auth.Authmode {
	case "Offline":
		return
	case "MojangAuth":
		//检测asTk是否有效
		ok, err := resp.Validate(auth.AsTk)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		if ok == true {
			return
		}
		authlogin(&auth.Account, &auth.Authserver)
	case "ThreeAuth":
		ygg.AuthURL = fmt.Sprintf("%s/authserver", auth.Authserver)
		bot.SessionURL = fmt.Sprintf("%s/sessionserver/session/minecraft/join", auth.Authserver)
		ok, err := resp.Validate(auth.AsTk)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		if ok == true {
			return
		}
		authlogin(&auth.Account, &auth.Authserver)
	default:
		log.Fatal("Unknown authmode")
		os.Exit(1)
	}
}

//利用auth更新配饰文件
func updateconfig() {
	//文件不存在
	_, err := os.Stat("conf.json")
	if os.IsNotExist(err) {
		conf.Players = append(conf.Players, auth)
		data, _ := json.Marshal(conf)
		var d1 = []byte(data)
		err := ioutil.WriteFile("conf.json", d1, 0666)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		return
	}
	//文件存在
	conf = loadconf()
	var i = 0
	//判断是否已存在该角色
	for _, v := range conf.Players {
		//判断验证模式
		if v.Authmode == "Offline" {
			//判断是否同一个账号
			if auth.Name == v.Name {
				return
			}
		} else {
			//判断是否同一个账号
			if auth.UUID == v.UUID {
				//判断是否需要更新AsTk
				if auth.AsTk != v.AsTk {
					break
				}
				return
			}
		}
		i++
	}
	if i == len(conf.Players) {
		conf.Players = append(conf.Players, auth)
	} else {
		conf.Players[i].AsTk = auth.AsTk
	}
	//替换配置文件
	data, _ := json.Marshal(conf)
	var d1 = []byte(data)
	if err := os.Remove("conf.json"); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err = ioutil.WriteFile("conf.json", d1, 0666)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func directlogin(uuid string) {
	log.Println("根据UUID读取配置")
	conf = loadconf()
	for _, v := range conf.Players {
		if v.UUID == uuid {
			log.Println("在配置文件中找到该角色")
			auth = v
			if auth.Authmode == "ThreeAuth" {
				ygg.AuthURL = fmt.Sprintf("%s/authserver", auth.Authserver)
				bot.SessionURL = fmt.Sprintf("%s/sessionserver/session/minecraft/join", auth.Authserver)
			}
			ok, err := resp.Validate(auth.AsTk)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			if ok == true {
				return
			}
			log.Println("无效的AsTk")
			authlogin(&auth.Account, &auth.Authserver)
			return
		}
	}
	log.Fatal("未记录的UUID，请使用-account添加登录")
}
