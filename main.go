package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"github.com/dablelv/go-huge-util/file"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

var cacheFileInfo sync.Map
var conf *Conf

func main(){
	filePath := flag.String("config","./conf.yaml","配置文件路径")
	newConf := flag.String("newconf","","生成配置文件路径")
	flag.Parse()

	if filePath == nil {
		tem := "./conf.yaml"
		filePath = &tem
	}

	if *newConf != "" {
		buf,err := yaml.Marshal(&Conf{
			Cmds:           []string{"go build"},
			Dir:            []string{"./"},
			TimeOut:        3000,
			GoPackageName:  "auto",
			FilterFileType: []string{"*"},
			Debug:          true,
		})
		if err != nil { log.Fatalln("yaml Marshal err:" ,err) }
		err = os.WriteFile(*newConf,buf,0666)
		if err != nil {
			log.Fatalln("write config file err:",err)
		}
		return
	}
	Load(*filePath)
	Run()


}

func Load(confPath ...string)  {
	var err error

	conf,err = NewConf(confPath...)

	if err != nil {
		log.Fatalln("read conf error :" + err.Error())
	}
	files,err := getFilePaths(conf.Dir,conf.FilterFileType)
	if err != nil {
		log.Fatalln(err)
	}
	for _, s := range files {
		if s[len(s)-3:] == "exe" {
			continue
		}
		setStore(s, gethash(s))
	}

	log.Printf(
		`reload dir: %v	reload timeOut:%v ms	go package name:%v	execute cmd:%v	filterFileType:%v	debug:%v`,
		conf.Dir,conf.TimeOut,conf.GoPackageName,conf.Cmds,conf.FilterFileType,conf.Debug)
}

func Run()  {

	for  {
		if ok,err := compare();!ok{
			if err != nil {
				panic(any(err))
			}
			time.Sleep(time.Millisecond*time.Duration(conf.TimeOut))
			continue
		}
		isRun := false
		for _, cmd := range conf.Cmds {
			cmd = strings.TrimLeft(strings.TrimRight(cmd," ")," ")
			switch cmd {
			case "go generate":
				cmd = cmd + " " + conf.GoPackageName
			case "go build":
				cmd = cmd + " -pkgdir " + conf.GoPackageName +" -o " + conf.GoPackageName +".exe"
			case "go run":
				cmd = cmd + " " + conf.GoPackageName
				isRun = true
			default:
				isRun = true
			}

			if conf.Debug { log.Println("execute:",cmd) }

			out,err := runCmd(cmd)
			if err != nil {
				log.Fatalln(err)
			}
			if conf.Debug { log.Println("execute result:",out) }else {fmt.Println(out)}
		}
		if !isRun {
			cmd := conf.GoPackageName +".exe"
			if conf.Debug { log.Println("execute:",cmd) }
			out,err := runCmd(cmd)
			if err != nil {
				log.Fatalln(err)
			}
			if conf.Debug {
				log.Println("execute result:",out)
			} else {
				fmt.Println(out)
			}
		}

	}


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

func compare() (bool,error)  {
	var wait sync.WaitGroup

	var isFresh bool

	files,err := getFilePaths(conf.Dir,conf.FilterFileType)
	if err != nil {
		return false, errors.New("Get All Files error :" + err.Error())
	}

	for _, _filePath := range files {
		if _filePath[len(_filePath)-3:] == "exe" {
			continue
		}
		wait.Add(1)
		go func(path string) {

			defer wait.Done()
			if isFresh {
				return
			}

			hax := gethash(path)
			loadHax := getStore(path)
			if loadHax == ""{
				setStore(path,hax)
				if conf.Debug { log.Println("new file:",path) }
				return
			}
			if loadHax != hax {
				setStore(path, hax)
				isFresh = true
				return
			}
		}(_filePath)

		if isFresh { break	}
	}

	wait.Wait()

	return isFresh,nil

}

func getStore(key string) string {
	val ,ok:= cacheFileInfo.Load(key)
	if !ok {
		return ""
	}

	return val.(string)
}

func setStore(key string,val string)  {
	if conf.Debug {
		fmt.Println("reload file ",key,val)
	}
	cacheFileInfo.Store(key,val)
}

func gethash(path string) (hash string) {
	buf,err := os.ReadFile(path)
	if err != nil {
		log.Println("open file err:",path,err)
		return  ""
	}

	Sha256 := sha256.New()
	Sha256.Write(buf)

	return hex.EncodeToString(Sha256.Sum(nil))

}

func getFilePaths(dirs []string,rules []string) ([]string,error) {

	newFiles := make([]string,0)
	for _, s := range dirs {
		files,err := file.GetDirAllEntryPathsFollowSymlink(s,false)
		if err != nil {
			return nil,err
		}
		files = filesFilter(files,rules)
		newFiles = append(newFiles, files...)
	}
	return newFiles,nil
}

func filesFilter(filePaths []string,rules []string) []string {

	var tempF = make([]string,0)
	var add bool

	for _, filePath := range filePaths {
		ext := path.Ext(filePath)
		add = true
		for _, rule := range rules {
			switch rule[0:1] {
			case ".":
				if rule != ext {
					add = false
				}
			case "!":
				if rule[1:2] == "." {
					if ext == rule[1:] {
						add = false
					}
				}else {
					if strings.Contains(filePath,rule[1:]) {
						add = false
					}
				}
			default:
				if rule[0:1] == "*" {
					add = true
				}else if !strings.Contains(filePath,rule) {
					add = false
				}
			}

			if !add {
				break
			}
		}

		if add {
			tempF = append(tempF, filePath)
		}
	}

	return tempF
}


