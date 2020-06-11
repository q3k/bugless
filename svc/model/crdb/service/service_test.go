package service

import (
	"context"
	"net"

	spb "github.com/q3k/bugless/proto/svc"

	log "github.com/inconshreveable/log15"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func dutModel() (spb.ModelClient, context.CancelFunc) {
	ctx, ctxC := context.WithCancel(context.Background())

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()

	d, err := inMemoryDatabase(ctx)
	if err != nil {
		panic(err)
	}
	if err := d.Migrate(); err != nil {
		panic(err)
	}
	spb.RegisterModelServer(s, &Service{
		db: d,
		l:  log.New("component", "service"),
	})

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := spb.NewModelClient(conn)

	go func() {
		<-ctx.Done()
		conn.Close()
		s.Stop()
	}()

	return client, ctxC
}
