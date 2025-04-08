package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/jingy/Go-Shortener/internal/storage"
	"github.com/jingy/Go-Shortener/pkg/shortener"
	pb "github.com/jingy/Go-Shortener/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedURLShortenerServer
	shortener *shortener.URLShortener
	storage   *storage.URLStorage
}

func (s *server) CreateShortURL(ctx context.Context, req *pb.CreateShortURLRequest) (*pb.CreateShortURLResponse, error) {
	// Create short URL using existing shortener
	shortCode, err := s.shortener.CreateShortURL(ctx, req.Url)
	if err != nil {
		return nil, err
	}

	// Get URL details from storage
	url, err := s.storage.GetURL(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return &pb.CreateShortURLResponse{
		ShortCode:  shortCode,
		ShortUrl:   url.ShortURL,
		CreatedAt:  url.CreatedAt.Unix(),
		ExpiresAt:  url.ExpiresAt.Unix(),
	}, nil
}

func (s *server) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	// Get URL from storage
	url, err := s.storage.GetURL(ctx, req.ShortCode)
	if err != nil {
		return nil, err
	}

	return &pb.GetOriginalURLResponse{
		OriginalUrl: url.OriginalURL,
		CreatedAt:   url.CreatedAt.Unix(),
		ExpiresAt:   url.ExpiresAt.Unix(),
	}, nil
}

func (s *server) GetURLStats(ctx context.Context, req *pb.GetURLStatsRequest) (*pb.GetURLStatsResponse, error) {
	// Get URL from storage
	url, err := s.storage.GetURL(ctx, req.ShortCode)
	if err != nil {
		return nil, err
	}

	// Get statistics (you'll need to implement this in your storage layer)
	stats, err := s.storage.GetURLStats(ctx, req.ShortCode)
	if err != nil {
		return nil, err
	}

	return &pb.GetURLStatsResponse{
		ShortCode:      req.ShortCode,
		TotalClicks:    stats.TotalClicks,
		UniqueVisitors: stats.UniqueVisitors,
		CreatedAt:      url.CreatedAt.Unix(),
		ExpiresAt:      url.ExpiresAt.Unix(),
		ClicksByCountry: stats.ClicksByCountry,
		ClicksByHour:    stats.ClicksByHour,
	}, nil
}

func main() {
	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	// Create DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Initialize storage
	urlStorage := storage.NewURLStorage(dynamoClient)
	counterStorage := storage.NewCounterStorage(dynamoClient)

	// Initialize shortener
	urlShortener := shortener.NewURLShortener(urlStorage, counterStorage)

	// Create gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterURLShortenerServer(s, &server{
		shortener: urlShortener,
		storage:   urlStorage,
	})

	// Register reflection service on gRPC server
	reflection.Register(s)

	log.Printf("Starting gRPC server on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
} 