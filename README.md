## ang admin

### vue
pnpm install
pnpm dev
pnpm build

### admin
go run main.go

## 项目介绍
ang-admin是ang全协议代理网关的管理端，包含管理后台接口和管理端web页面，主要是用来管理用户配置数据，代理数据，与ang引擎交互，发送配置到ang，ang基于给定配置运行
采用restful风格
基于github.com/bugfan/rest路由框架
基于xorm数据库框架
暂时基于sqlite3本地数据库
管理端web页面要始终支持中英双语
管理端web页面要始终支持手机/PC双端响应式布局
管理端web原地址仓库 https://github.com/pure-admin/vue-pure-admin 

## ai agent注意事项
改完代码测试运行之后一定要自己停到程序，不要影响我启动测试代码

## todo
按增量下发配置
tunnel client链接时候暂时没有限制，谁都能链接，这个应该在tunnel server添加/编辑时候搞一个配置项，控制是否允许未注册的tunnel client token能链接上来
tunnel菜单里的server/client改成tunnel和node是否更好
server.json tunnel部分，也许不需要制定sni了，因为都是一份自签发证书，内置了