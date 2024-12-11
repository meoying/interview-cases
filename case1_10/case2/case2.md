### 索引失效-命中数据量占比大

#### 初始化数据
运行case2_test.go文件下的TestCase2方法。

#### 观察sql
索引扫描的行记录占比少，正常走索引

EXPLAIN select * from `orders` where uid = 123 and create_time > '2024-10-25 18:47:59';
![img_1.png](img_1.png)



索引扫描的行记录占比大，走全表

EXPLAIN select * from `orders` where uid = 123 and create_time > '2024-09-13 18:47:59';
![img_2.png](img_2.png)

