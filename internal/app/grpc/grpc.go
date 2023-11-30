package grpc

import (
	"context"
	"errors"
	"log"
	"net"
	"net/url"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/app"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/auth"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/internal/config"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/models"
	"github.com/Yandex-Practicum/go-musthave-shortener-trainer/pkg/shortener"
)

type Server struct {
	shortener.UnimplementedShortenerServer
	Server   *grpc.Server
	instance *app.Instance
}

func (s *Server) Shorten(ctx context.Context, request *shortener.ShortenRequest) (*shortener.ShortenResponse, error) {
	u, err := url.Parse(request.Url)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, app.ErrParseURL.Error())
	}
	shorten, err := s.instance.Shorten(ctx, u)
	if err != nil {
		return nil, err
	}
	resp := shortener.ShortenResponse{Result: shorten}
	return &resp, nil
}

func (s *Server) BatchShorten(ctx context.Context, req *shortener.BatchShortenRequest) (*shortener.BatchShortenResponse, error) {
	var batch []models.BatchShortenRequest
	for _, r := range req.Batch {
		batchReq := models.BatchShortenRequest{
			CorrelationID: r.CorrelationId,
			OriginalURL:   r.OriginalUrl,
		}
		batch = append(batch, batchReq)
	}

	shorten, err := s.instance.BatchShorten(batch, ctx)
	if errors.Is(err, app.ErrParseURL) {
		return nil, status.Errorf(codes.InvalidArgument, "Cannot parse given string as URL")
	}

	if errors.Is(err, app.ErrURLLength) {
		return nil, status.Errorf(codes.Internal, "invalid shorten URLs length")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	var resp []*shortener.BatchResponse
	for _, u := range shorten {
		resp = append(resp, &shortener.BatchResponse{CorrelationId: u.CorrelationID, ShortUrl: u.ShortURL})
	}

	return &shortener.BatchShortenResponse{
		Result: resp,
	}, nil
}
func (s *Server) BatchRemove(_ context.Context, req *shortener.BatchRemoveRequest) (*emptypb.Empty, error) {
	id, err := uuid.FromString(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot convert user id")
	}
	go func() {
		s.instance.RemoveChan <- models.BatchRemoveRequest{UID: id, Ids: req.Ids}
	}()
	return &emptypb.Empty{}, nil
}
func (s *Server) Statistics(ctx context.Context, req *shortener.StatisticsRequest) (*shortener.StatisticsResponse, error) {
	//  можно использовать для получения IP адреса
	//p, _ := peer.FromContext(ctx)
	statistics, err := s.instance.Statistics(ctx, req.Ip)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "Denied")
	}
	return &shortener.StatisticsResponse{Urls: uint32(statistics.Urls), Users: uint32(statistics.Users)}, nil
}
func (s *Server) Expand(ctx context.Context, req *shortener.UrlRequest) (*shortener.UrlResponse, error) {
	loadURL, err := s.instance.LoadURL(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &shortener.UrlResponse{OriginalUrl: loadURL.String()}, nil
}
func (s *Server) UserUrls(ctx context.Context, req *shortener.UserUrlsRequest) (*shortener.UserUrlsResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	log.Println(md)
	id, err := uuid.FromString(req.Uuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot convert user id")
	}
	userContext := auth.Context(ctx, id)
	users, err := s.instance.LoadUsers(userContext)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}
	var urls []*shortener.UserUrls
	for _, u := range users {
		urls = append(urls, &shortener.UserUrls{OriginalUrl: u.OriginalURL, ShortUrl: u.ShortURL})
	}
	return &shortener.UserUrlsResponse{Urls: urls}, nil
}
func (s *Server) Ping(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.instance.Ping(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return empty, nil
}

func NewShortenerServer(instance *app.Instance) *Server {
	//port := helper.GetEnvValue("RPC_SERVER_PORT", "50051")
	logrus.Printf("Starting gRPC server on port: %v", config.GrpcPort)
	lis, err := net.Listen("tcp", config.GrpcPort)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	server := &Server{instance: instance}
	shortener.RegisterShortenerServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	server.Server = s
	return server
}
