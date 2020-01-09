# goDDNS
利用Dnspod和阿里云的API中转，实现DDNS的功能
可以用来给mikrotik routeros使用

Using Dnspod and Alibaba Cloud API relay to implement DDNS functions
Can be used for mikrotik routeros

# install
`go install`

# Mikrotik 脚本
```
#PPPoE
:local pppoe "pppoe-out1"

#DDNS Token
:local token "XXX"

#DDNS param
:local record "XXX"
:local domain "XXX.com"

:global ddnsIPROS
:local ipnew [/ip address get [/ip address find interface=$pppoe] address]
:set ipnew [:pick $ipnew 0 ([len $ipnew] -3)]
:if ($ipnew != $ddnsIPROS) do={
    :log error "ddns update start"
    :local url "https://XXX.XXX.com/dnspod\?token=$token&ip=$ipnew&domain=$domain&record=$record"
    :local result [/tool fetch url=$url http-method=get as-value output=user]
    :if ($result->"status" = "finished") do={
        :if ($result->"data" = "1") do={
            :log error ("ddns update ok IP:" . $ipnew)
            :set ddnsIPROS $ipnew
        } else={
            :log error "ddns update error"
        }
    } else {
        :log error "ddns fetch error"
    }
} else {
    :log error "ddns do not need change"
}
```
