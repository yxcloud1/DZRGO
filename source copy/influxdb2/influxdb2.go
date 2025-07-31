package influxdb2

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	token  string
	bucket string
	org    string
	host   string
)

// 设置 InfluxDB2 的连接选项
func SetOption(_host string, _token string, _bucket string, _org string) {
	token = _token
	bucket = _bucket
	org = _org
	host = _host
}

// 写入数据点
func Write(table string, tags map[string]string, fields map[string]interface{}, ts time.Time) error {
	// 创建客户端
	client := influxdb2.NewClient(host, token)
	defer client.Close()

	// 获取写入 API
	writeAPI := client.WriteAPIBlocking(org, bucket)

	// 构造 point
	p := influxdb2.NewPoint(
		table,
		tags,
		fields,
		ts,
	)

	// 写入 point
	return writeAPI.WritePoint(context.Background(), p)
}