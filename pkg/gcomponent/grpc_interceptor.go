/**
 * Create Time:2023/11/13
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package gcomponent

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/qionggemens/gcommon/pkg/gentity"
	"github.com/qionggemens/gcommon/pkg/glog"
	util "github.com/qionggemens/gcommon/pkg/gutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	maxBodyLen = 1024
)

func getBodyStr(body interface{}) string {
	bodyStr := fmt.Sprintf("%+v", body)
	if len(bodyStr) > maxBodyLen {
		return bodyStr[:maxBodyLen]
	}
	return bodyStr
}

func getMdOfClient(ctx context.Context) (context.Context, metadata.MD) {
	md, exists := metadata.FromIncomingContext(ctx)
	var traceId string
	if !exists {
		traceId = strconv.FormatInt(time.Now().UnixMicro(), 10)[4:]
	} else {
		arr := md.Get(gentity.MdKeyTraceId)
		if len(arr) == 0 {
			traceId = strconv.FormatInt(time.Now().UnixMicro(), 10)[4:]
		} else {
			traceId = arr[0]
		}
	}

	outMd, exists := metadata.FromOutgoingContext(ctx)
	if !exists {
		outMd = metadata.Pairs(gentity.MdKeyTraceId, traceId)
		return metadata.NewOutgoingContext(ctx, outMd), outMd
	}
	copied := outMd.Copy()
	copied.Set(gentity.MdKeyTraceId, traceId)
	return metadata.NewOutgoingContext(ctx, copied), copied
}

func getMdOfServer(ctx context.Context) metadata.MD {
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		return metadata.MD{}
	}
	copied := md.Copy()
	if len(copied.Get(gentity.MdKeyTraceId)) == 0 {
		copied.Set(gentity.MdKeyTraceId, "")
	}
	return copied
}

// GrpcServerInterceptor 服务端拦截器
func GrpcServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rsp interface{}, err error) {
	reqStr := getBodyStr(req)
	clientAddr := util.GetGrpcClientAddr(ctx)
	md := getMdOfServer(ctx)
	defer func() {
		if p := recover(); p != nil {
			glog.Errorf("[GRPC SERVER] %s fail [From:%s] - md:%+v, req:%s, err:%v, stack:%s", info.FullMethod, clientAddr, md, reqStr, p, string(debug.Stack()))
			rsp = nil
			err = status.Errorf(codes.Internal, "panic: %v", p)
		}
	}()
	glog.Infof("[GRPC SERVER] %s begin [From:%s] - md:%+v, req:%s", info.FullMethod, clientAddr, md, reqStr)
	bt := time.Now()
	rsp, err = handler(ctx, req)
	cost := time.Since(bt).Milliseconds()
	if err != nil {
		glog.Errorf("[GRPC SERVER] %s fail [From:%s] - cost:%dms, md:%+v, req:%s, msg:%s", info.FullMethod, clientAddr, cost, md, reqStr, err.Error())
	} else {
		rspStr := getBodyStr(rsp)
		glog.Infof("[GRPC SERVER] %s success [From:%s] - cost:%dms, md:%+v, req:%s, rsp:%s", info.FullMethod, clientAddr, cost, md, reqStr, rspStr)
	}
	return rsp, err
}

// GrpcClientInterceptor 客户端拦截器
func GrpcClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	reqStr := getBodyStr(req)
	serverAddr := util.GetGrpcClientAddr(ctx)
	outCtx, md := getMdOfClient(ctx)
	defer func() {
		if p := recover(); p != nil {
			glog.Errorf("[GRPC CLIENT] %s fail [To:%s] - md:%+v, req:%s, err:%v, stack:%s", method, serverAddr, md, reqStr, p, string(debug.Stack()))
			err = status.Errorf(codes.Internal, "panic: %v", p)
		}
	}()
	glog.Infof("[GRPC CLIENT] %s begin [To:%s] - md:%+v, req:%s", method, serverAddr, md, reqStr)
	bt := time.Now()
	err = invoker(outCtx, method, req, reply, cc, opts...)
	cost := time.Since(bt).Milliseconds()
	if err != nil {
		glog.Errorf("[GRPC CLIENT] %s fail [To:%s] - cost:%dms, md:%+v, req:%s, msg:%s", method, serverAddr, cost, md, reqStr, err.Error())
	} else {
		rspStr := getBodyStr(reply)
		glog.Infof("[GRPC CLIENT] %s success [To:%s] - cost:%dms, md:%+v, req:%s, rsp:%s", method, serverAddr, cost, md, reqStr, rspStr)
	}
	return err
}
