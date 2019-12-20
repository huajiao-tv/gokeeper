package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huajiao-tv/gokeeper/model"
	pb "github.com/huajiao-tv/gokeeper/pb/go"
	"google.golang.org/grpc"
)

func main() {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, "127.0.0.1:7001", opts...)
	if err != nil {
		panic(err)
	}

	client := pb.NewSyncClient(conn)
	c, err := client.Sync(ctx)
	if err != nil {
		panic(err)
	}

	str, _ := json.Marshal(model.NodeInfo{})
	pbEventReq := &pb.ConfigEvent{
		EventType: pb.ConfigEventType_CONFIG_EVENT_NONE,
		Data:      string(str),
	}
	err = c.Send(pbEventReq)
	if err != nil {
		fmt.Println("send error:", err)
	} else {
		fmt.Println("send successfully")
	}

	pbEventResp, err := c.Recv()
	fmt.Println("recv:", pbEventResp, err)

}
