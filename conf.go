package main

import (
	"bytes"
	"errors"
	"github.com/dablelv/go-huge-util/file"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
)

type Conf struct {
	fileName string
	Cmds []string	`yaml:"cmds"`
	Dir []string	`yaml:"dirs"`
	TimeOut int		`yaml:"timeOut"`
	GoPackageName string	`yaml:"goPackageName"`
	FilterFileType []string	`yaml:"filterFileType"`
	Debug bool	`yaml:"debug"`
}

func NewConf(path ...string) (*Conf,error) {
	if path == nil { path = []string{"./conf.yaml"} }
	buf,err := os.ReadFile(path[0])
	if err != nil {
		log.Println("read conf error :open ./conf.yaml: The system cannot find the file specified.")
	}
	var conf Conf
	err = yaml.Unmarshal(buf,&conf)
	if err != nil {
		return nil,err
	}
	if len(conf.Dir) == 0 {
		conf.Dir = []string{"./"}
	}
	if conf.TimeOut == 0 {
		conf.TimeOut = 3000
	}
	if len(conf.Cmds) == 0 {
		conf.Cmds = []string{"go build"}
	}
	if conf.GoPackageName == "" || conf.GoPackageName == "auto"{
		name ,err := getProjectName("./")
		if err != nil {
			log.Fatalln(any("GO package 加载错误：" + err.Error()))
		}

		conf.GoPackageName = strings.TrimLeft(name," ")
	}
	conf.fileName = path[0]
	return &conf,nil
}

func getProjectName(path ...string) (name string,Err error) {
	if path == nil { path= []string{"./"} }
	files, err := file.GetDirAllEntryPathsFollowSymlink(path[0], false)
	if err != nil {
		return "search error at current directory ",err
	}

	for _, s := range files {
		if s == "go.mod"  || strings.Contains(s,"go.mod") {
			buf,err := os.ReadFile(s)
			if err != nil {
				Err = errors.New("open and read go.mod err:"+err.Error())
				continue
			}

			if lf := bytes.Index(buf,[]byte("module"));lf > -1 {
				if lf2 := bytes.IndexByte(buf[lf:],'\n');lf2 > 0 {
					Err = nil
					name = string(bytes.ReplaceAll(buf[lf+6:lf2],[]byte{13},[]byte{}))
					break
				}
				Err = errors.New("Project name could not be found at go.mod ")
			}

		}
	}

	return
}