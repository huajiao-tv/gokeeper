// gsync：
// google grpc sync server
// 采用google grpc协议封装的sync接口

package sync

import (
	"github.com/huajiao-tv/gokeeper/model"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"github.com/huajiao-tv/gokeeper/server/logger"
)

type GSyncServer struct{}

var (
	defaultModelEventResp = &model.Event{EventType: model.EventNone}
)

//grpc协议同步配置
//目前只要报错，就直接把当前stream连接断开，后续需要修改
//@todo 规范grpc error code
func (server *GSyncServer) Sync(syncServer pb.Sync_SyncServer) error {
	for {
		pbEventReq, err := syncServer.Recv()
		if err != nil {
			return err
		}
		modelEventReq, err := model.ParseEvent(pbEventReq)
		if err != nil {
			logger.Logex.Trace("GSync", "ParseEvent", err.Error(), pbEventReq)
			return err
		}
		modelEventResp, node, err := eventProxy(modelEventReq)
		// add prometheus info
		AddPromNodeEvent("req", modelEventResp, node)
		if err != nil {
			logger.Logex.Trace("GSync", "eventProxy", err.Error(), modelEventReq)
			return err
		}
		// grpc stream Send nil指针的时候，会直接报错 "error while marshaling: proto: Marshal called with nil"，所以不能send nil
		if modelEventResp == nil {
			modelEventResp = defaultModelEventResp
		}
		pbEventResp, err := model.FormatEvent(modelEventResp)
		if err != nil {
			logger.Logex.Trace("GSync", "FormatEvent", err.Error(), modelEventResp)
			return err
		}
		err = syncServer.Send(pbEventResp)
		if err != nil {
			logger.Logex.Trace("GSync", "syncServer.Send", err.Error(), pbEventResp)
			return err
		}
	}
}
