# go-cache
go语言实现的缓存组件

功能点：
1.LRU实现
2.Pending实现，避免缓存击穿


设计亮点：
1.对外暴露的API足够简单实用
2.pengding的实现逻辑和底层的lru实现逻辑解耦，可以自行替换底层实现
3.pending包本身只实现pending的逻辑。【cacheSize,expireTime,threadSafe】由底层的cache自行保证


整体设计思路:
1.通过chan来实现对某个key的pending，避免缓存击穿

补充知识：
缓存穿透：缓存和数据库中都没有的数据
缓存击穿：缓存中没有但数据库中有的数据


缓存观察指标：
1.缓存命中率
2.缓存占用的内存(例子：100万结构体缓存400M)【需要更具业务情况看，大部分是可以全量缓存的】
3.缓存数据的有效性