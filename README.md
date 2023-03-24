# esdump

一个简单es导出cli程序,可以实现从 Elasticsearch 中导出数据到CSV文件中。

## 安装

```shell
go install github.com/lifei6671/esdump
```

## 使用

```shell
$ esdump -h

GLOBAL OPTIONS:
   --query value, -q value                                                Query filename in Lucene syntax.
   --match-all value, -m value [ --match-all value, -m value ]            List of <field>:<direction> pairs to filter.
   --output-file value, -o value                                          CSV file location. [required]
   --es-server value, -e value [ --es-server value, -e value ]            Elasticsearch host URL. (default: "http://localhost:9200")
   --auth value, -a value                                                 Elasticsearch basic authentication in the form of username:password.
   --es-version value, -E value                                           Elasticsearch version (default: "v7")
   --index-prefixes value, -i value [ --index-prefixes value, -i value ]  Index name prefix(es). Default is ['logstash-*']. (default: "log-*")
   --fields value, -f value [ --fields value, -f value ]                  List of selected fields in output.
   --sort value, -s value [ --sort value, -s value ]                      List of <field>:<desc|asc> pairs to sort on.
   --page-size value, -p value                                            Maximum number returned per page. (default: 1000)
   --scroll-size value, -S value                                          Scroll size for each batch of results.  (default: 5m0s)
   --range-field value, -R value                                          scope field for query (default: "@timestamp")
   --range-value value, -V value [ --range-value value, -V value ]        List of <field>:<direction> pairs to range on. (default: "2023-03-23T20:53:34.0097493+08:00", "2023-03-24T20:53:34.0449937+08:00")
   --raw-query value, -r value                                            Switch query format in the Query DSL.
   --ignore-err, -n                                                       Ignore non-fatal error messages. (default: true)
   --debug                                                                Debug mode on. (default: true)
   --help, -h                                                             show help
   --version, -v                                                          print the version


```

## 示例

### 参数

| 参数                                     | 说明                                                                                                          |
|----------------------------------------|-------------------------------------------------------------------------------------------------------------|
| [-q -query](#query)                    | 指定一个ES的DSL查询文件路径 , 需要以@开头： `@~/home/work/dsl.txt`                                                           |
| [-m --match-all](#match-all)           | 指定一个简单的查询语句     , 查询语句需要以冒号分割，第一个为查询的索引名，第二段查询索引值：`json.api:/user/query`                                    |
| [-o --output-file](#output-file)       | 指定输出的文件路径   , 需要确保有些权限:`/home/work/output.csv`                                                              |
| [-e --es-server](#es-server)           | 指定ES的服务器地址  ,多个节点以`,`分隔:`http://127.0.0.1:9200,http://127.0.0.1:9201`                                       |
| [-a --auth](#auth)                     | ES服务器认证账户  , 仅支持账号密码认证:`username:password`                                                                  |
| [-E --es-version](#es-version)         | ES服务器版本             ,  目前仅支持v7版本:`v7`                                                                       |
| [-i --index-prefixes](#index-prefixes) | 指定要查询的索引前缀       ,多个字段需要以`,`分隔：`json.api,json.name,json.id`                                                 |    
| [-s --sort](#sort)                     | 指定排序字段                    , 多个排序字段以`,`分隔:`@timestamp:desc,id:asc`                                           |
| [-p --page-size](#page-size)           | 指定每次查询的数据条数          , 请设置一个合理值:`1000`                                                                      |
| [-S --scroll-size](#scroll-size)       | 指定滚动查询时快照保持时间    , `@timestamp`                                                                             |
| [-V --range-value](#range-value)       | 指定ES查询的`range`值,以`:`分隔的值，第一个字为小于等于，第二个值为大于:`2023-03-20:2023-03-21`                                          |
| [-r --raw-query](#raw-query)           | 原始的ES查询语句，对于较短的查询语句，可以指定该参数:`{"size": 1000,"sort":[{"_doc":"asc"}],"query":{"match_all":{}},"scroll":"1m"}` |
| [-n --ignore-err](#ignore-err)         | 指定是否忽略非致命异常，当出现查询失败，或解析失败时不中断查询，而是继续查询剩余结果:`true`                                                           |
| [--debug](#debug)                      | 开启Debug模式，会打印更多日志                                                                                           |
| [-v --version](#version)               | 打印程序版本                                                                                                      |


### 示例

#### <a id="query"></a>query

如果查询的是一个复杂的语句，可以通过该参数指定DSL查询语句所在文件，执行是会自动加载该文件作为ES的查询语句：

```shell
esdump --query=@~/home/work/dsl.txt
```

#### <a id="match-all"></a>match-all

用于指定一个简单的查询语句，当`queyr` 和 `match-all` 都传时，以`query`为最高优先级。字段的格式需要以`:`分隔，多个查询条件以`,`分隔。

```shell
esdump --match-all=json.uri:/user/query 
```

#### <a id="output-file"></a>output-file

```shell
esdump --match-all=json.uri:/user/query --output-file=/home/work/output.csv
```