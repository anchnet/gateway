package g

import (
	"fmt"
)

type TransferResp struct {
	Msg        string
	Total      int
	ErrInvalid int
	Latency    int64
}

func (t *TransferResp) String() string {
	s := fmt.Sprintf("TransferResp total=%d, err_invalid=%d, latency=%dms",
		t.Total, t.ErrInvalid, t.Latency)
	if t.Msg != "" {
		s = fmt.Sprintf("%s, msg=%s", s, t.Msg)
	}
	return s
}

type ProxyMetricValue struct {
	Endpoint  string      `json:"device_id"`
	Metric    string      `json:"metric"`
	Value     interface{} `json:"value"`
	Step      int64       `json:"step"`
	Tags      string      `json:"tags"`
	Timestamp string      `json:"timestamp"`
}

func (this *ProxyMetricValue) String() string {
	return fmt.Sprintf(
		"<Endpoint:%s, Metric:%s,Tags:%s, Step:%d, Time:%s, Value:%v>",
		this.Endpoint,
		this.Metric,
		this.Tags,
		this.Step,
		this.Timestamp,
		this.Value,
	)
}

type ProxyResp struct {
	Msg   string `json:"message"`
	Satus string `json:"status"`
}

func (p *ProxyResp) String() string {
	return fmt.Sprintf("ProxyResp Msg=%s, Satus=%s", p.Msg, p.Satus)
}
