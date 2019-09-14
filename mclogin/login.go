package mclogin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/AlecAivazis/survey"
	"github.com/Tnze/go-mc/bot"
	ygg "github.com/Tnze/go-mc/yggdrasil"
)

//Player is a profile of aplayer.
type Player struct {
	Name       string `json:"name"`
	UUID       string `json:"UUID"`
	AsTk       string `json:"AsTk"`
	Account    string `json:"account"`
	Authserver string `json:"authserver"`
	Authmode   string `json:"authmode"`
}

//Config used to restore config in the program
type Config struct {
	Players []Player `json:"players"`
}

var (
	conf Config
	resp *ygg.Access
)

//Authlogin is used when you need to login with passwd
func Authlogin(account, authserver *string, auth *Player) {
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
		var (
			no          int
			preSelected []string
			selected    string
		)
		//把安排survey包的单选中的选项所需要的切片搞出来
		for _, v := range resp.AvailableProfiles() {
			preSelected = append(preSelected, fmt.Sprintf("[%d]", no)+v.Name)
			no++
		}
		//让玩家进行选择
		prompt := &survey.Select{
			Message: "选择一个角色:",
			Options: preSelected,
		}
		survey.AskOne(prompt, &selected)
		//获得选中的用户的序号
		fmt.Sscanf(selected, "[%d]", &no)
		//refresh刷新获得该用户AsTk
		if err = resp.Refresh(&resp.AvailableProfiles()[no]); err != nil {
			log.Fatal(err)
			os.Exit(1)

		}
	}
	//安排登陆账户
	auth.UUID, auth.Name = resp.SelectedProfile()
	auth.AsTk = resp.AccessToken()
	return
}

//Loadconf is used to load Config from file
func Loadconf() Config {
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

//LoadConfiglogin can load config and ask you to choose a profile and login
func LoadConfiglogin(auth *Player) {
	var conf = Loadconf()
	log.Println("成功读取配置")
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
		Message: "选择一个账户:",
		Options: preSelected,
	}
	survey.AskOne(prompt, &selected)
	fmt.Sscanf(selected, "[%d]", &selectedNo)
	*auth = conf.Players[selectedNo]
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
		Authlogin(&auth.Account, &auth.Authserver, auth)
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
		Authlogin(&auth.Account, &auth.Authserver, auth)
	default:
		log.Fatal("未知验证模式")
		os.Exit(1)
	}
}

//AddtoConfig is used to update the config file with a provided profile
func AddtoConfig(auth *Player) {
	//文件不存在
	_, err := os.Stat("conf.json")
	if os.IsNotExist(err) {
		conf.Players = append(conf.Players, *auth)
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
	conf = Loadconf()
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
		conf.Players = append(conf.Players, *auth)
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

//Directlogin is used to load config and login according to the uuid
func Directlogin(uuid string, auth *Player) {
	log.Println("根据UUID读取配置")
	conf = Loadconf()
	for _, v := range conf.Players {
		if v.UUID == uuid {
			log.Println("在配置文件中找到该角色")
			*auth = v
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
			Authlogin(&auth.Account, &auth.Authserver, auth)
			return
		}
	}
	log.Fatal("未记录的UUID，请使用-account添加登录")
}

//Configrm means remove-mode
func Configrm() {
	conf := Loadconf()
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
		Message: "选择你要删除的账户:",
		Options: preSelected,
	}
	survey.AskOne(prompt, &selected)
	fmt.Sscanf(selected, "[%d]", &selectedNo)
	conf.Players = append(conf.Players[:selectedNo], conf.Players[selectedNo+1:]...)
	data, _ := json.Marshal(conf)
	var d1 = []byte(data)
	if err := os.Remove("conf.json"); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	err := ioutil.WriteFile("conf.json", d1, 0666)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	name := false
	prompt1 := &survey.Confirm{
		Message: "已删除选择账户，继续删除账户？",
	}
	survey.AskOne(prompt1, &name)
	if name == true {
		Configrm()
	}
	os.Exit(1)
}
