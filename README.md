# multi-deploy
基于rsync的多机代码发布工具

## 起因
在使用Jenkins部署代码时候发现多机部署有点麻烦,每次新增或释放机子都需要修改配置文件.
在此痛点下,为了提高效率,本工具就诞生了.

## 特性
- Go语言开发，编译简单、运行高效
- 支持网络连通性检测
- 支持协程方式同步
- 支持带参运行
- 支持钉钉通知

## 准备
本工具底层使用rsync进行同步,使用前请确保rsync可用,并确保本机通过秘钥可直接SSH到目标机器

本工具还依赖Redis,使用前请确保Redis可用,并做以下配置
1. 新建LIST类型名为'key'配置项的值 eg:multi-deploy-hosts
2. 把目标主机的IP地址填入LIST,建议填写内网地址 eg:172.16.100.210

## 使用
```sh
# clone
git clone https://github.com/cnlyzy/multi-deploy.git
cd multi-deploy

# build
go mod download
go build
chmod +x multi-deploy

# config
mv conf/config.example.json conf/config.json
vim conf/config.json

# run
./multi-deploy -exclude -p=/data/www/project
```

## 参数
```sh
-p 代码的绝对路径(不带'/')
-exclude 同步代码需要排除的文件或目录,一般为日志文件等,使用该参数请确保项目目录存在'excludeFrom'配置项同名文件
-v 显示当前版本并退出
```

## TODO
- [ ] 日志文件记录
- [ ] Windows支持

> 欢迎Star,Fork,PR
