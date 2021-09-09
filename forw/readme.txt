## 介绍
国密ssl 四层加密转发代理

## 命令
解压，sm2Certs文件夹保持同级目录

./forw.linux -l :7765 -lm gmssl -f :7766 -fm raw

-l： 监听地址
-lm: 监听的数据报文格式 gmssl, 或者raw
-f：转发地址
-fm: 转发的数据报文格式，gmssl, 或者raw

示例：
PSI分server端和client端，server监听7766端口，走socket通信，forw代理端口1088
server起代理
cd
client起代理
./forw.linux -l :7766 -lm raw -f 172.16.88.77:1088 -fm gmssl





