package rpc

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	pfc "github.com/niean/goperfcounter"
	cmodel "github.com/open-falcon/common/model"
	cutils "github.com/open-falcon/common/utils"

	"github.com/anchnet/gateway/g"
	"github.com/anchnet/gateway/sender"
)

type Transfer int

func (this *Transfer) Ping(req cmodel.NullRpcRequest, resp *cmodel.SimpleRpcResponse) error {
	return nil
}

func (t *Transfer) Update(args []*cmodel.MetricValue, reply *g.TransferResp) error {
	return RecvMetricValues(args, reply, "rpc")
}

// process new metric values
func RecvMetricValues(args []*cmodel.MetricValue, reply *g.TransferResp, from string) error {
	start := time.Now()
	reply.ErrInvalid = 0

	items := []*cmodel.MetaData{}
	for _, v := range args {
		if v == nil {
			reply.ErrInvalid += 1
			continue
		}

		// 历史遗留问题.
		// 老版本agent上报的metric=kernel.hostname的数据,其取值为string类型,现在已经不支持了;所以,这里硬编码过滤掉
		if v.Metric == "kernel.hostname" {
			reply.ErrInvalid += 1
			continue
		}

		if v.Metric == "" || v.Endpoint == "" {
			reply.ErrInvalid += 1
			continue
		}

		if v.Type != g.COUNTER && v.Type != g.GAUGE && v.Type != g.DERIVE {
			reply.ErrInvalid += 1
			continue
		}

		if v.Value == "" {
			reply.ErrInvalid += 1
			continue
		}

		if v.Step <= 0 {
			reply.ErrInvalid += 1
			continue
		}

		if len(v.Metric)+len(v.Tags) > 510 {
			reply.ErrInvalid += 1
			continue
		}

		errtags, tags := cutils.SplitTagsString(v.Tags)
		if errtags != nil {
			reply.ErrInvalid += 1
			continue
		}

		// TODO 呵呵,这里需要再优雅一点
		now := start.Unix()
		if v.Timestamp <= 0 || v.Timestamp > now*2 {
			v.Timestamp = now
		}

		fv := &cmodel.MetaData{
			Metric:      v.Metric,
			Endpoint:    v.Endpoint,
			Timestamp:   v.Timestamp,
			Step:        v.Step,
			CounterType: v.Type,
			Tags:        tags, //TODO tags键值对的个数,要做一下限制
		}

		valid := true
		var vv float64
		var err error

		switch cv := v.Value.(type) {
		case string:
			vv, err = strconv.ParseFloat(cv, 64)
			if err != nil {
				valid = false
			}
		case float64:
			vv = cv
		case int64:
			vv = float64(cv)
		default:
			valid = false
		}

		if !valid {
			reply.ErrInvalid += 1
			continue
		}

		fv.Value = vv
		items = append(items, fv)
	}

	// statistics
	cnt := int64(len(items))
	pfc.Meter("Recv", cnt)
	if from == "rpc" {
		pfc.Meter("RpcRecv", cnt)
	} else if from == "http" {
		pfc.Meter("HttpRecv", cnt)
	}

	cfg := g.Config()
	if cfg.Transfer.Enabled {
		sender.Push2SendQueue(items)
	}

	reply.Msg = "ok"
	reply.Total = len(args)
	reply.Latency = (time.Now().UnixNano() - start.UnixNano()) / 1000000

	return nil
}

// 匹配代理服务器数据上报
func RecvProxyMetricValues(args []*g.ProxyMetricValue, reply *g.TransferResp, from string) error {
	start := time.Now()
	reply.ErrInvalid = 0

	items := []*cmodel.MetaData{}
	timeLayout := "2006-01-02 15:04:05"
	for _, v := range args {
		if v == nil {
			reply.ErrInvalid += 1
			continue
		}

		if v.Metric == "" || v.Endpoint == "" {
			reply.ErrInvalid += 1
			continue
		}

		if v.Value == "" {
			reply.ErrInvalid += 1
			continue
		}

		if v.Step <= 0 {
			reply.ErrInvalid += 1
			continue
		}

		if len(v.Metric)+len(v.Tags) > 510 {
			reply.ErrInvalid += 1
			continue
		}

		// 这个校验应该不需要，验证tag格式
		errtags, tags := cutils.SplitTagsString(v.Tags)
		if errtags != nil {
			reply.ErrInvalid += 1
			continue
		}

		//timeStep传过来的是string类型
		now := start.Unix()
		loc, _ := time.LoadLocation("Local")
		ts, _ := time.ParseInLocation(timeLayout, v.Timestamp, loc)
		valueTimeStep := ts.Unix()
		if valueTimeStep <= 0 || valueTimeStep > now*2 {
			valueTimeStep = now
		}

		// change metric and endpoint
		v.Endpoint = "_server_" + v.Endpoint

		valid := true
		var vv float64
		var err error

		switch cv := v.Value.(type) {
		case string:
			if strings.Contains(strings.ToLower(v.Metric), "disk") {
				r, _ := regexp.Compile("[a-zA-Z]*$")
				valUnit := r.FindString(v.Value.(string))
				if valUnit != "" {
					vSelf := strings.Replace(v.Value.(string), valUnit, "", 1)
					vv, err = strconv.ParseFloat(vSelf, 64)
					if err != nil {
						valid = false
					}
					if valid {
						switch strings.ToLower(valUnit) {
						case "bytes":
							continue
						case "m":
							vv = vv * 1024 * 1024
						case "mb":
							vv = vv * 1024 * 1024
						case "g":
							vv = vv * 1024 * 1024 * 1024
						case "gb":
							vv = vv * 1024 * 1024 * 1024
						case "t":
							vv = vv * 1024 * 1024 * 1024 * 1024
						case "tb":
							vv = vv * 1024 * 1024 * 1024 * 1024
						default:
							valid = false
						}
					}
				}
			} else if strings.Contains(strings.ToLower(v.Metric), "status") {
				vv, err = strconv.ParseFloat(v.Value.(string), 64)
				if vv == 0 {
					vv = 1
				} else {
					vv = -1
				}
			}
		case float64:
			if strings.Contains(strings.ToLower(v.Metric), "status") {
				println(cv)
				if cv == 0 {
					vv = 1
				} else {
					vv = -1
				}
			} else {
				vv = cv
			}
		case int64:
			if strings.Contains(strings.ToLower(v.Metric), "status") {
				if cv == 0 {
					vv = 1
				} else {
					vv = -1
				}
			} else {
				vv = float64(cv)
			}
		default:
			valid = false
		}

		if !valid {
			reply.ErrInvalid += 1
			continue
		}

		thirdMetrics := g.Config().ThirdMetrics
		if b, ok := thirdMetrics[v.Metric]; ok && b != "" {
			v.Metric = thirdMetrics[v.Metric]
		}

		fv := &cmodel.MetaData{
			Metric:      v.Metric,
			Endpoint:    v.Endpoint,
			Timestamp:   valueTimeStep,
			Step:        v.Step,
			CounterType: g.GAUGE,
			Tags:        tags,
			Value:       vv,
		}

		println(fv.String())
		// fv.Value = vv
		items = append(items, fv)
	}

	// statistics
	cnt := int64(len(items))
	pfc.Meter("Recv", cnt)
	if from == "rpc" {
		pfc.Meter("RpcRecv", cnt)
	} else if from == "http" {
		pfc.Meter("HttpRecv", cnt)
	}

	cfg := g.Config()
	if cfg.Transfer.Enabled {
		sender.Push2SendQueue(items)
	}

	reply.Msg = "ok"
	reply.Total = len(args)
	reply.Latency = (time.Now().UnixNano() - start.UnixNano()) / 1000000

	return nil
}
