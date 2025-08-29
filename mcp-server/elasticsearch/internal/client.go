package internal

type ElasticsearchClient struct {
	BaseURL string
}

func NewElasticsearchClient() (*ElasticsearchClient, error) {
	// 从配置文件获取Elasticsearch配置
	// 这里可以根据需要添加具体的配置项
	return &ElasticsearchClient{
		BaseURL: "http://localhost:9200", // 默认地址，可以从配置文件读取
	}, nil
}
