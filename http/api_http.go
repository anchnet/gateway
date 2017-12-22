package http

import (
	"encoding/json"
	"net/http"

	cmodel "github.com/open-falcon/common/model"

	"github.com/open-falcon/gateway/g"
	trpc "github.com/open-falcon/gateway/receiver/rpc"
)

func configApiHttpRoutes() {
	http.HandleFunc("/api/push", func(w http.ResponseWriter, req *http.Request) {
		if req.ContentLength == 0 {
			http.Error(w, "blank body", http.StatusBadRequest)
			return
		}

		decoder := json.NewDecoder(req.Body)
		var metrics []*cmodel.MetricValue
		err := decoder.Decode(&metrics)
		if err != nil {
			http.Error(w, "decode error", http.StatusBadRequest)
			return
		}

		reply := &g.TransferResp{}
		trpc.RecvMetricValues(metrics, reply, "http")
		package http

		import (
			"encoding/json"
			"io"
			"io/ioutil"
			"net/http"
			"strings"
		
			cmodel "github.com/open-falcon/common/model"
		
			"github.com/anchnet/gateway/g"
			trpc "github.com/anchnet/gateway/receiver/rpc"
		)
		
		//原生agent push 数据
		func configApiHttpRoutes() {
			http.HandleFunc("/api/push", func(w http.ResponseWriter, req *http.Request) {
				if req.ContentLength == 0 {
					http.Error(w, "blank body", http.StatusBadRequest)
					return
				}
		
				decoder := json.NewDecoder(req.Body)
				var metrics []*cmodel.MetricValue
				err := decoder.Decode(&metrics)
				if err != nil {
					http.Error(w, "decode error", http.StatusBadRequest)
					return
				}
		
				reply := &g.TransferResp{}
				trpc.RecvMetricValues(metrics, reply, "http")
		
				RenderDataJson(w, reply)
			})
		}
		
		// 代理服务器push监控数据，需要特殊处理
		func configApiProxyHttpRoutes() {
			http.HandleFunc("/api/proxy/push", func(w http.ResponseWriter, req *http.Request) {
				if req.ContentLength == 0 {
					http.Error(w, "blank body", http.StatusBadRequest)
					return
				}
		
				decoder := json.NewDecoder(req.Body)
				var metrics []*g.ProxyMetricValue
				err := decoder.Decode(&metrics)
				if err != nil {
					http.Error(w, "decode error", http.StatusBadRequest)
					return
				}
		
				reply := &g.TransferResp{}
				trpc.RecvProxyMetricValues(metrics, reply, "http")
		
				RenderDataJson(w, reply)
			})
		}
		
		// 同步用户接口
		func configApiAddUser() {
			http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				if req.ContentLength == 0 {
					http.Error(w, "Request Error", http.StatusBadRequest)
					return
				}
				reply := &g.ProxyResp{}
				proxyHttpReq(req, reply)
				RenderDataJson(w, reply)
			})
		}
		
		func proxyHttpReq(req *http.Request, reply *g.ProxyResp) {
			reply.Satus = "error"
			reply.Msg = ""
			cli := &http.Client{}
			smartEyeAddr := g.Config().SmartEye
			proxyAddr := smartEyeAddr + req.URL.Path
			body := make([]byte, req.ContentLength)
			_, err := io.ReadFull(req.Body, body)
			if err != nil {
				reply.Msg = "proxy check reqbody error.error info:" + err.Error()
				return
			}
			proxyReq, err := http.NewRequest(req.Method, proxyAddr, strings.NewReader(string(body)))
			if err != nil {
				reply.Msg = "proxy init request error.error info:" + err.Error()
				return
			}
			proxyReq.Header.Set("Content-Type", req.Header.Get("Content-Type"))
			proxyRes, err := cli.Do(proxyReq)
			if err != nil {
				reply.Msg = "proxy do request error.error info:" + err.Error()
				return
			}
		
			defer proxyRes.Body.Close()
		
			body2, err := ioutil.ReadAll(proxyRes.Body)
			// io.ReadFull(proxyRes.Body, body2)
			// println("current response:" + string(body2))
		
			// decoder := json.NewDecoder(proxyRes.Body)
			err = json.Unmarshal(body2, &reply)
			if err != nil {
				reply.Msg = "proxy decode resbody error.error info:" + err.Error()
				return
			}
		}
				RenderDataJson(w, reply)
	})
}
