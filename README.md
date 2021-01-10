# FishBot
Minecraft内自动钓鱼  
Fishing in Minecraft automatically

基于[https://github.com/Tnze/go-mc](https://github.com/MscBaiMeow/FishBotReload)

# 这个项目是我Golang学习的处女作，已经修修补补一段时间了，这次go-mc有一个较大的更新，是时候重写了
# 另见 [https://github.com/MscBaiMeow/FishBotReload](https://github.com/MscBaiMeow/FishBotReload)

#### 使用方法

------

  -account string
​        Mojang账号

  -auth string
​        验证服务器（外置登陆） (default "https://authserver.mojang.com")

  -ip string
​        服务器IP (default "localhost")

  -name string
​        游戏ID

  -passwd string
​        Mojang账户密码

  -port int
​        端口，默认25565 (default 25565)

  -realms
​        加入领域服

  -t int
​        自动重新抛竿时间 (default 45)


如果你正确打开的话，接下来会有指示的

#### 注意

------

~~不支持SRV解析，请自行解析~~

目前优先解析SRV

具体方法GUI版本的readme看去

<https://github.com/Amazefcc233/FishBotGUI>

（懒得写了

#### 鸣谢

------

[Tnze](https://github.com/Tnze)（go-mc项目作者，FishBot原作者）

[fcc](https://github.com/Amazefcc233)（GUI版本作者，提供测试服务器）

[Jing332](https://github.com/jing332)（领域服支持项目抓包）

[Miaoscraft](https://miaoscraft.cn)（感谢相遇）
