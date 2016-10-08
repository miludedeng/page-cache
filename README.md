## 页面静态化中间件
-------
可应用于动态动态服务器与代理服务器之间，经过此插件可将页面静态化，如果已存在静态页面，则直接返回。如果没有，则会生成静态页面并返回
#### 使用说明:
###### 参数
`-port` <br/>服务端口

`-proxy`  <br/>是动态服务器的域名或ip, 需要以http:// 开头

`-concatcss`<br/> 可选true/false,该选项将开启将会把页面中直接引用的css直接加入页面中，以减少请求次数

`-redishost` <br/>redis地址

`-redisport` <br/>redis 端口

`-redisdb` <br/>redis database

`-maxidle` <br/>最大的空闲连接数，表示即使没有redis连接时依然可以保持N个空闲的连接，而不被清除，随时处于待命状态。默认值1

`-maxactive` <br/>最大的激活连接数，表示同时最多有N个连接。默认1000

###### 注意事项
*关于过期时间，不同的访问路径可以通过Header的EXPDATE设置

以nginx为例<br/>
　　在location中添加 proxy_set_header EXPDATE 180;<br/> 
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;*注意此处时间只能以秒为单位

*如果页面在开发或测试时，可以在浏览器的url后拼接
`nocache=true`，该url则不会使用缓存

#### Docker方式启动示例：
`docker  run -d --name pagestatic --link redis:redis -v /opt/page_static/conf:/usr/local/pagestatic/conf -v /opt/page_static/logs:/usr/local/pagestatic/logs -p 3000:3000 pagecache`