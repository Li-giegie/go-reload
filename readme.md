
## go-reload 是一个热刷新Go程序的工具，可以在一定程度上简化运行操作步骤，支持在此基上二次开发新的项目，几乎不用看源码，在结构体对象上的Run函数中支持传递一个回调，回调参数是一个[]string 为发生改变的路径文件
### 特点：
    （1）.简单 完全可以0配置来运行在你的项目中 为你热刷新
    （2）.快速 基于go语言 性能无需多言

### 使用方式：
    (1) git clone https://github.com/Li-giegie/go-reload.git
    
    (2) 构建可执行程序 go build

    (3) 将可执行文件目录 添加到系统环境变量Path中

    (4) 在任意项目中通过可执行文件名运行
[环境变量设置教程]   (https://blog.csdn.net/zhengqijun_/article/details/53222301)

### 配置文件：
    # 执行的命令 默认执行 go build 可以是多条命令同时执行
    cmds: []
    # 检测的热刷新的目录或文件 多选项 支持
    dirs: []
    # 检测间隔时间单位毫秒
    timeOut: 0
    # go package name ：auto为自动检测 也可以自己手动输入
    goPackageName: ""
    # 那些文件会被检测 
    # ["*"] 检测所有文件  [.go] 检测.go结尾的文件 [!.go] 检测非.go结尾的文件 | [main.go] 检测main.go | [！main.go] 检测非main.go的文件
    filterFileType: []
    # 是否开启调试模式 打印详细的运行信息
    debug: false

## go-reload 命令列表
    -config [文件路径] 指定配置文件 运行
    -newconf [文件路径] 生成默认配置文件

## 