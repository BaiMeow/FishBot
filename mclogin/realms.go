package mclogin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/realms"
)

//Checkrealms can get IP of the realm so that you can login in realms
func Checkrealms(ip *string, port *int, c *bot.Client, version *string, auth *Player) error {
	var r *realms.Realms
	r = realms.New(*version, c.Name, c.AsTk, c.Auth.UUID)
	servers, err := r.Worlds()
	if err != nil {
		return err
	}
	var i, no = 0, 0
	//list realms

	//agree TOS
	if err := r.TOS(); err != nil {
		return err
	}
	var (
		selected    string
		preSelected []string
		selectedNo  int
	)
	for _, v := range servers {
		preSelected = append(preSelected, fmt.Sprintf("[%d]", i)+v.Name)
		i++
	}
	prompt := &survey.Select{
		Message: "Choose a realm:",
		Options: preSelected,
	}
	survey.AskOne(prompt, &selected)
	fmt.Sscanf(selected, "[%d]", &selectedNo)
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
