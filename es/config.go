package es

import "time"

type Config struct {
	Query      string        `json:"query"`
	Filename   string        `json:"filename"`
	EsServer   []string      `json:"es_server"`
	EsVersion  string        `json:"es_version"`
	Auth       string        `json:"auth"`
	Index      []string      `json:"index"`
	MatchAll   []string      `json:"match_all"`
	Fields     []string      `json:"fields"`
	Sort       []string      `json:"sort"`
	Scroll     time.Duration `json:"scroll"`
	RawQuery   string        `json:"raw_query"`
	Debug      bool          `json:"debug"`
	MaxSize    int           `json:"max_size"`
	RangeValue []string      `json:"range_value"`
	RangeField string        `json:"range_field"`
	IgnoreErr  bool          `json:"ignore_err"`
}
