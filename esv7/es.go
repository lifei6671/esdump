package esv7

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"

	"github.com/lifei6671/esdump/es"
)

type esClient struct {
	esClient *elastic.Client
	conf     es.Config
}

func (e *esClient) Dump(ctx context.Context, receive chan<- json.RawMessage) error {
	queryReader, err := e.query()
	if err != nil {
		//如果编译查询语句失败直接返回
		return err
	}
	var scrollId string
	var scrollIdList []string
	// 初始化搜索请求
	req := esapi.SearchRequest{
		Index:  e.conf.Index,
		Body:   queryReader,
		Scroll: e.conf.Scroll,
	}
	// 执行搜索请求，获取第一批结果和滚动 ID
	res, err := req.Do(ctx, e.esClient)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	defer func() {
		// 发送清除滚动请求以释放资源
		clearReq := esapi.ClearScrollRequest{
			ScrollID: scrollIdList,
		}
		clearRes, err := clearReq.Do(context.Background(), e.esClient)
		if err != nil {
			log.Error().Strs("scroll_ids", scrollIdList).Err(err)
			return
		}
		_ = clearRes.Body.Close()
		if clearRes.IsError() {
			log.Error().Strs("scroll_ids", scrollIdList).Strs("warning", clearRes.Warnings()).Send()
		}
	}()
	//读取所有返回值
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	gResult := gjson.ParseBytes(b)
	//获取当前文档的总数
	total := int64(gResult.Get("hits.total.value").Float())

	scrollId = gResult.Get("_scroll_id").String()
	//获取滚动上下文id
	log.Debug().Str("scroll_id", scrollId).Int64("total", total).Send()
	scrollIdList = append(scrollIdList, scrollId)
	//写完了一定要关闭chan
	defer close(receive)

	//获取最终查询的结果
	hits := gResult.Get("hits.hits.#._source").Array()
	for _, hit := range hits {
		receive <- json.RawMessage(hit.Raw)
	}
	// 持续滚动直到没有更多的文档需要返回
	for scrollId != "" {
		scrollId, err = e.scroll(ctx, scrollId, e.conf.Scroll, receive)
		if err != nil {
			log.Error().Str("scroll_id", scrollId).Err(err).Send()
			if !e.conf.IgnoreErr {
				return err
			}
		}
		scrollIdList = append(scrollIdList, scrollId)
	}
	return nil
}

func (e *esClient) query() (io.Reader, error) {
	var queryReader io.Reader
	if len(e.conf.RawQuery) > 0 {
		log.Debug().Any("query", e.conf.RawQuery).Send()
		queryReader = esutil.NewJSONReader(e.conf.RawQuery)
	} else {
		query := map[string]any{
			"query": map[string]any{
				"match_all": map[string]any{},
			},
		}

		//如果传入的是一个文件地址，则从文件中序列化查询参数
		if strings.HasPrefix(e.conf.Query, "@") {
			body, err := os.ReadFile(strings.TrimPrefix(e.conf.Query, "@"))
			if err != nil {
				log.Error().Str("filename", e.conf.Query).Err(err)
				return nil, err
			}
			err = json.Unmarshal(body, &query)
			if err != nil {
				log.Error().Str("filename", e.conf.Query).Err(err)
				return nil, err
			}
		} else {
			match := map[string]any{}
			if len(e.conf.MatchAll) > 0 {
				for _, tag := range e.conf.MatchAll {
					fields := strings.Split(tag, ":")
					if len(fields) == 2 {
						match[fields[0]] = fields[1]
					}
				}

			}
			rangeFilter := map[string]map[string]any{}
			if len(e.conf.RangeValue) > 0 && len(e.conf.RangeField) > 0 {
				rangeFilter[e.conf.RangeField] = map[string]any{}
				if len(e.conf.RangeValue) > 0 {
					rangeFilter[e.conf.RangeField]["gte"] = e.conf.RangeValue[0]
				}
				if len(e.conf.RangeValue) > 1 {
					rangeFilter[e.conf.RangeField]["lt"] = e.conf.RangeValue[1]
				}
			}
			query["query"] = map[string]any{
				"bool": map[string]any{
					"must": []any{
						map[string]any{
							"match": match,
						},
						map[string]any{
							"range": rangeFilter,
						},
					},
				},
			}
		}

		if len(e.conf.Fields) > 0 {
			log.Debug().Strs("fields", e.conf.Fields).Send()
			query["_source"] = e.conf.Fields
		}
		if len(e.conf.Sort) > 0 {
			var sorts []map[string]string
			for _, sort := range e.conf.Sort {
				sortItem := strings.Split(sort, ":")
				if len(sortItem) == 2 {
					sorts = append(sorts, map[string]string{
						sortItem[0]: sortItem[1],
					})

				}
			}
			query["sort"] = sorts
		}

		queryReader = esutil.NewJSONReader(query)
		log.Debug().Any("query", query).Send()
	}
	return queryReader, nil
}

func (e *esClient) scroll(ctx context.Context, scrollID string, scrollTimeout time.Duration, receive chan<- json.RawMessage) (string, error) {
	req := esapi.ScrollRequest{
		ScrollID: scrollID,
		Scroll:   scrollTimeout,
	}
	//一直遍历请求
	res, err := req.Do(ctx, e.esClient)
	if err != nil {
		log.Error().Str("scroll_id", scrollID).Err(err).Send()
		return "", err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	gResult := gjson.ParseBytes(b)

	scrollID = gResult.Get("_scroll_id").String()

	log.Debug().Msg(gResult.Get("hits.hits").Raw)

	hits := gResult.Get("hits.hits.#._source").Array()
	if len(hits) == 0 {
		return "", nil
	}
	for _, hit := range hits {
		receive <- json.RawMessage(hit.Raw)
	}
	return scrollID, nil
}

func NewClient(c es.Config) (es.Client, error) {
	esConfig := elastic.Config{
		Addresses: c.EsServer,
	}
	auth := strings.Split(c.Auth, ":")
	if len(auth) == 2 {
		esConfig.Username = auth[0]
		esConfig.Password = auth[1]
	}
	client, err := elastic.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	return &esClient{
		esClient: client,
		conf:     c,
	}, nil
}
