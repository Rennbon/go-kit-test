package main

import (
	"context"
	"github.com/Rennbon/donself/pb"
	"google.golang.org/grpc"
	"log"
	"time"
)

func main() {
	ctx1, _ := context.WithTimeout(context.Background(), time.Second*5)
	address := "www.rennbon.online:10690"
	address = "localhost:10690"
	conn, err := grpc.DialContext(ctx1, address, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Println("did not connect: %v", err)
		return
	}
	defer conn.Close()
	c := pb.NewDoneselfClient(conn)
	req := &pb.AllMyTargetsRequest{
		PageIndex: 1,
		PageSize:  10,
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*10)
	for i := uint32(0); i < 1000; i++ {
		time.Sleep(time.Millisecond * 250)
		go func(num uint32) {
			req.PageIndex = num
			res, err := c.AllMyTargets(ctx, req)
			if err != nil {
				log.Println("num:", i, err)
				//os.Exit(1)
			} else {
				log.Println("num:", i, res)
			}
		}(i)
	}
}
