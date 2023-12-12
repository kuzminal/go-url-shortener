package grpc

import (
	"context"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
	"github.com/gofrs/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor перехватчик для проверки наличия пользователя и генерации его если он отсутствует
func AuthInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	var uid *uuid.UUID
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	a := md.Get("auth")
	if len(a) > 0 {
		uid, _ = auth.DecodeUIDFromHex(a[0])
	}

	if uid == nil {
		userID := ensureRandom()
		uid = &userID
	}

	value, err := auth.EncodeUIDToHex(*uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot encode auth")
	}
	md.Append("auth", value)

	ctx = auth.Context(ctx, *uid)
	ctx = metadata.NewIncomingContext(ctx, md)

	err = grpc.SendHeader(ctx, md)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func ensureRandom() (res uuid.UUID) {
	for i := 0; i < 10; i++ {
		res = uuid.Must(uuid.NewV4())
	}
	return
}
