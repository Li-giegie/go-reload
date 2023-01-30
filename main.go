package main

import (
	"bytes"
	"flag"
	"fmt"
	go_scout "github.com/Li-giegie/go-scout"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main(){

	confPath := initFlag()

	conf,err := newConf(confPath)
	if err != nil {
		log.Fatalln(err)
	}

	sock,_,err := go_scout.New(1000,conf.Dir...)
	if err != nil {
		log.Fatalln(err)
	}

	err = sock.Scout(func(changePath []*go_scout.FileInfo) {

		if len(changePath) <= 2 {
			if len(changePath)>1 && changePath[0].Name == conf.GoPackageName + ".exe" || changePath[1].Name == conf.GoPackageName + ".exe"{
				return
			}
		}

		for _, cmd := range conf.Cmds {
			result,err := runCmd(cmd)
			if err != nil {
				fmt.Println("执行命令：",cmd,"错误：",err)
			}
			fmt.Println(strings.TrimLeft(strings.TrimRight(result,"\n"),"\n"))
		}

	})
}

func initFlag()  string {
	confPath := flag.String("config","./conf.yaml","配置文件路径")
	newConfPath := flag.String("newconf","","生成配置文件路径")
	flag.Parse()
	fmt.Println("newConfPath ",newConfPath)
	if *newConfPath != ""{
		createConf(*newConfPath)
		os.Exit(0)
	}
	return *confPath
}

func runCmd(cmdStr string) (string,error){
	list := strings.Split(cmdStr, " ")
	cmd := exec.Command(list[0],list[1:]...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stderr.String(),err
	}

	return out.String(),nil
}

func createConf(path string)  {
	buf,err := yaml.Marshal(&_Conf{
		Cmds:           []string{"go build"},
		Dir:            []string{"./"},
		TimeOut:        3000,
		GoPackageName:  "auto",
		FilterFileType: []string{"*"},
		Debug:          true,
	})
	if err != nil { log.Fatalln("yaml Marshal err:" ,err) }

	err = os.WriteFile(path,buf,0666)
	if err != nil {
		log.Fatalln("write config file err:",err)
	}

}


