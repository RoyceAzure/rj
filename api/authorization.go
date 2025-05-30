package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/RoyceAzure/rj/api/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

type IAuthorizer[T token.UserIDConstraint] interface {
	AuthorizUser(ctx context.Context) (*token.Payload[T], string, error)
}

type Authorizer[T token.UserIDConstraint] struct {
	tokenMaker token.Maker[T]
}

func NewAuthorizor[T token.UserIDConstraint](tokenMaker token.Maker[T]) IAuthorizer[T] {
	return &Authorizer[T]{
		tokenMaker: tokenMaker,
	}
}

/*
從ctx 取得metadata
從metadata取得authHeader => 檢查格式 => 檢查Bearer => 用tokenMaker 檢查token 內容
*/
func (auth *Authorizer[T]) AuthorizUser(ctx context.Context) (*token.Payload[T], string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, "", fmt.Errorf("misssing metadata")
	}

	values := md.Get(authorizationHeaderKey)
	if len(values) == 0 {
		return nil, "", fmt.Errorf("misssing authorization header")
	}
	authHeader := strings.Fields(values[0])
	if len(authHeader) != 2 {
		return nil, "", fmt.Errorf("invalid auth format")
	}
	authType := strings.ToLower(authHeader[0])
	if authType != authorizationTypeBearer {
		return nil, "", fmt.Errorf("unsportted authorization type : %s", authType)
	}

	accessToken := authHeader[1]
	payload, err := auth.tokenMaker.VertifyToken(accessToken)
	if err != nil {
		return nil, "", fmt.Errorf("invaliad access token : %s", err)
	}
	return payload, accessToken, nil
}
