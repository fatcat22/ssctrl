
This is a tool that help you to set local proxy and control [Shadowsocks2](https://github.com/shadowsocks/go-shadowsocks2) conveniently.

[Shadowsocks2](https://github.com/shadowsocks/go-shadowsocks2) is a very great project, it's simple but very useful. Furthermore, it is wrote by go. But when using Shadowsocks2 I have a lot of work to do manually, such as set up proxy config of my OS, make Shadowsocks2 start when I log into my computer, and so on. In addition, I have to set up a PAC server manually when using PAC(Proxy auto-config). So I write this tool to help me do those things.


# Build

Just run `go build` in the project directory(But make sure your `go` support `go mod` command)


# Install & Usage

1. build this project
2. build [Shadowsocks2](https://github.com/shadowsocks/go-shadowsocks2). (The output executable file name should be `go-shadowsocks2` or `go-shadowsocks2.exe` on winodws)
3. move executable files generated by the two prjects to the same directory
4. make sure ~/.ssctrl(C:\Users\yourname\.ssctrl on windows) directory exist
5. copy gfwlist.js(locate at this project directory) to ~/.ssctrl/gfwlist.js
6. run ssctrl in your terminal

`ssctrl` should be running in your system after those steps, but before you can use proxy normally, you should tell `ssctrl` the config of your proxy server and enable it with http API:
```
# NOTE: The default port of http API is 1083, 

# set proxy server
curl -X POST "127.0.0.1:1083/updateServers" -d '{"serverName":{"address":"1.1.1.1","port":"1111","crypto":"AEAD_AES_128_GCM","password":"mypassword1"}}'

# set the server you'd like to use
curl -X POST "127.0.0.1:1083/currentServer" -d "serverName"

# set proxy mode. The default mode is pac, so if you'd like 
#   to use pac mode you could omit this step
curl -X POST "127.0.0.1:1083/mode" -d "pac"

# enable proxy
curl -X GET "127.0.0.1:1083/enable"
```

Now you should using your proxy normally. If you'd like to `ssctrl` works automatically, you can set `ssctrl` auto-run when you log into your OS:
```
curl -X POST "127.0.0.1:1083/autorun" -d "enable"
```

See more API information at [API](#API) section below.


# API

The default port of http API server is 1083.

### enable and disable proxy

> curl -X GET "127.0.0.1:1083/enable"
> curl -X GET "127.0.0.1:1083/disable"


### get config
> curl -X GET "127.0.0.1:1083/config"

return value on success:
>{"enabled":true,"mode":"pac","localPort":"1080","pacPort":"1082","apiPort":"1083","usingServer":"myserver1","servers":{"myserver1":{"address":"1.1.1.1","port":"8488","crypto":"AEAD_CHACHA20_POLY1305","password":"yourpwd"}}}


### exit

> curl -X POST "127.0.0.1:1083/exit"

`exit` will shut `ssctrl` process down.


### change mode

> curl -X POST "127.0.0.1:1083/mode" -d "NewMode"

`NewMode` should be `pac` or `global`


### change local port

> curl -X POST "127.0.0.1:1083/localPort" -d "1088"

NOTE: The proxy will not work if you change the default local port when you are using PAC mode. I will fix it in the future.


### change pac port

> curl -X POST "127.0.0.1:1083/pacPort" -d "2088"


### change api port

> curl -X POST "127.0.0.1:1083/apiPort" -d "3088"


### change current server

> curl -X POST "127.0.0.1:1083/currentServer" -d "server_name"


### add/change server(s) config

> curl -X POST "127.0.0.1:1083/updateServers" -d '{"server1Name":{"address":"1.1.1.1","port":"1111","crypto":"AEAD_AES_128_GCM","password":"mypassword1"},"server2Name":{"address":"1.1.1.1","port":"1111","crypto":"","password":"mypassword1"}}'

As you see, you can set more than one server's information when post `updateServers` command.


### remove server(s) config

> curl -X POST "127.0.0.1:1083/removeServers" -d '["server1Name", "server2Name"]'

You can remove more than one server's config when post 'removeServers' command.


### set autorun 

> curl -X POST "127.0.0.1:1083/autorun" -d "enable"
> curl -X POST "127.0.0.1:1083/autorun" -d "disable"


# TODO

- [ ] api `autorun` with disable will cause process exit when process is running as a service.
- [ ] re-generate pac file
- [ ] osapi on windows
- [ ] pac whitelist and blacklist


# reference

http://findproxyforurl.com/pac-functions/
https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_(PAC)_file
https://raw.githubusercontent.com/breakwa11/gfw_whitelist/master/whitelist.pac