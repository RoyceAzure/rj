package util

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type contextKey string

const (
	grcpGateWayUserAgentHeader            = "grpcgateway-user-agent"
	userAgentHeader                       = "usesr-agent"
	xForwardedForHeader                   = "x-forwarded-for"
	RequestIDKey               contextKey = "X-Request-ID"
)

type MetaData struct {
	UserAgent string `json:"user_agent"`
	ClientIP  string `json:"client_ip"`
	RequestId string `json:"request_id"`
}

func ExtractMetaData(ctx context.Context) *MetaData {
	mtda := &MetaData{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			mtda.UserAgent = userAgents[0]
		}
		if userAgents := md.Get(grcpGateWayUserAgentHeader); len(userAgents) > 0 {
			mtda.UserAgent = userAgents[0]
		}
		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			mtda.UserAgent = clientIPs[0]
		}

		if request_id := md.Get(string(RequestIDKey)); len(request_id) > 0 {
			mtda.RequestId = request_id[0]
		}

		if p, ok := peer.FromContext(ctx); ok {
			mtda.ClientIP = p.Addr.String()
		}
	}
	return mtda
}

func NewOutGoingMetaData(ctx context.Context, accessToken string) context.Context {
	orimd := ExtractMetaData(ctx)
	md := metadata.New(map[string]string{
		string(RequestIDKey): orimd.RequestId,
	})
	if accessToken != "" {
		md["authorization"] = []string{fmt.Sprintf("%s %s", "bearer", "authorization_payload")}
	}
	return metadata.NewOutgoingContext(ctx, md)
}
