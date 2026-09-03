/**
 * Create Time:2024/1/12
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package gcomponent

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/qionggemens/gcommon/pkg/glog"
)

// HttpServer http
type HttpServer struct {
	Interceptors []HandleInterceptor
	RequestMap   map[string]func(rw http.ResponseWriter, req *http.Request)
}

// HandleInterceptor 拦截器
type HandleInterceptor interface {
	PreHandle(rw http.ResponseWriter, req *http.Request) bool
	PostHandle(rw http.ResponseWriter, req *http.Request)
}

func (server *HttpServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	urlPath := req.URL.Path
	defer func() {
		if e := recover(); e != nil {
			glog.Errorf("[HTTP ACCESS] handle fail - path:%s, err:%v, stack:%s", urlPath, e, string(debug.Stack()))
			// 尽量返回 500；若业务已写过 Header 可能无效，但不把 panic 伪装成成功
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}()
	glog.Infof("[HTTP ACCESS] req begin - path:%s", urlPath)
	bt := time.Now()
	reqMapping, isOk := server.RequestMap[urlPath]
	if !isOk {
		glog.Errorf("[HTTP ACCESS] req end - path:%s, msg:404", req.URL.Path)
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	for i := 0; i < len(server.Interceptors); i++ {
		interceptor := server.Interceptors[i]
		isOk = interceptor.PreHandle(rw, req)
		if !isOk {
			glog.Errorf("[HTTP ACCESS] req end - path:%s, interceptor:%v, msg:preHandle fail", req.URL.Path, interceptor)
			return
		}
	}
	reqMapping(rw, req)
	for i := len(server.Interceptors) - 1; i >= 0; i-- {
		interceptor := server.Interceptors[i]
		interceptor.PostHandle(rw, req)
	}
	glog.Infof("[HTTP ACCESS] req end - cost:%dms, path:%s", time.Since(bt).Milliseconds(), req.URL.Path)
}
