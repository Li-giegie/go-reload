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

type reload struct {
	conf *Conf
	cacheFileInfo sync.Map
}

//
func New(confPath ...string)  *reload {
	var err error
	var _reload reload
	_reload.conf,err = NewConf(confPath...)

	if err != nil {
		log.Fatalln("read conf error :" + err.Error())
	}
	files,err := getFilePaths(_reload.conf.Dir,_reload.conf.FilterFileType)
	if err != nil {
		log.Fatalln(err)
	}

	for _, s := range files {

		_reload.cacheFileInfo.Store(s,getHash(s))
	}

	log.Printf(
		`reload dir: %v	reload timeOut:%v ms	go package name:%v	execute cmd:%v	filterFileType:%v	debug:%v`,
		_reload.conf.Dir,_reload.conf.TimeOut,_reload.conf.GoPackageName,_reload.conf.Cmds,_reload.conf.FilterFileType,_reload.conf.Debug)

	return &_reload

}

func main(){
	filePath := flag.String("config","./conf.yaml","配置文件路径")
	val := flag.String("newconf","","生成配置文件路径")
	flag.Parse()
	createConf(val)

	if filePath == nil {
		tem := "./conf.yaml"
		filePath = &tem
	}

	l := New(*filePath)
	l.Run()

}

func (r *reload) Run(changeFiles ...func(cf []string))  {

	for  {
		ok,_changeFiles,err := r.compare();
		if !ok{
			if err != nil {
				panic(any(err))
			}
			time.Sleep(time.Millisecond*time.Duration(r.conf.TimeOut))
			continue
		}

		if changeFiles != nil {
			changeFiles[0](_changeFiles)
		}

		if r.conf.Debug {
			log.Println("编辑的文件：",_changeFiles)
		}

		var isRun bool
		for _, cmd := range r.conf.Cmds {
			cmd = strings.TrimLeft(strings.TrimRight(cmd," ")," ")
			switch cmd {
			case "go generate":
				cmd = cmd + " " + r.conf.GoPackageName
			case "go build":
				cmd = cmd + " -pkgdir " + r.conf.GoPackageName +" -o " + r.conf.GoPackageName +".exe"
			case "go run":
				cmd = cmd + " " + r.conf.GoPackageName
				isRun = true
			default:
				isRun = true
			}

			if r.conf.Debug { log.Println("execute:",cmd) }

			out,err := runCmd(cmd)
			if err != nil {
				log.Fatalln(err)
			}
			if out != "" {
				if r.conf.Debug { log.Println("execute result	[\n",out,"\n]") }else {fmt.Println(out)}
			}
		}
		if !isRun {
			cmd := r.conf.GoPackageName +".exe"
			if r.conf.Debug { log.Println("execute:",cmd) }
			out,err := runCmd(cmd)
			if err != nil {
				log.Fatalln(err)
			}
			if r.conf.Debug {
				log.Println("execute result	[\n",out,"\n]")
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

func (r *reload) compare() (bool,[]string,error)  {
	var wait sync.WaitGroup
	var _changeFiles = make([]string,0)
	var isFresh bool

	files,err := getFilePaths(r.conf.Dir,r.conf.FilterFileType)
	if err != nil {
		return false,nil, errors.New("Get All Files error :" + err.Error())
	}

	for _, _filePath := range files {

		wait.Add(1)
		go func(filepath string) {
			defer wait.Done()

			hax := getHash(filepath)
			loadHax := r.getStore(filepath)
			if loadHax == ""{
				r.setStore(filepath,hax)
				if r.conf.Debug { log.Println("new file:",filepath) }
				return
			}
			if loadHax != hax {
				r.setStore(filepath, hax)
				if path.Ext(filepath) == ".exe" || path.Base(filepath) == path.Base(r.conf.GoPackageName){
					if r.conf.Debug {
						log.Println("可执行文件已更新...",path.Base(filepath))
					}
					return
				}

				if strings.TrimLeft(path.Base(filepath) ,`\`) == strings.TrimLeft(path.Base(r.conf.fileName),`\`) {
					log.Println("conf 文件已更改")
					fn := r.conf.fileName
					r.conf ,err = NewConf(fn)
					if err != nil {
						log.Fatalln(fn," :配置文件修改格式存在错误 程序已停止运行：",err)
					}
				}

				_changeFiles = append(_changeFiles, filepath)
				isFresh = true
				return
			}
		}(_filePath)

	}

	wait.Wait()

	return isFresh,_changeFiles,nil

}

func (r *reload) getStore(key string) string {
	val ,ok:= r.cacheFileInfo.Load(key)
	if !ok {
		return ""
	}

	return val.(string)
}

func (r *reload) setStore(key string,val string)  {
	if r.conf.Debug {
		fmt.Println("change file ",key,val)
	}
	r.cacheFileInfo.Store(key,val)
}

func getHash(path string) (hash string) {
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

func createConf(newConf *string)  {
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
		os.Exit(0)
	}

}


