#### docker 编译

```
docker run --rm -it -v ~/myapp/goapp/src:/go/src golang:1.9.4-alpine 
```

进入容器 编译 
```
go build -o guul main.go

```


在当前宿主机目录下得到一个guul的可执行文件

构建APP镜像即可



