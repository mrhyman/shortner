package grpc

import (
	"context"
	"errors"

	"github.com/mrhyman/shortner/internal/model"
	"github.com/mrhyman/shortner/internal/service"
	"github.com/mrhyman/shortner/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ShortenerHandler struct {
	grpcapi.UnimplementedShortenerServiceServer
	urlService *service.URLService
}

func NewShortenerHandler(urlService *service.URLService) *ShortenerHandler {
	return &ShortenerHandler{
		urlService: urlService,
	}
}

func (h *ShortenerHandler) ShortenURL(ctx context.Context, req *grpcapi.URLShortenRequest) (*grpcapi.URLShortenResponse, error) {
	shortURL, err := h.urlService.Shorten(ctx, req.GetUrl())
	if err != nil {
		if err == model.ErrInvalidURL {
			return nil, status.Error(codes.InvalidArgument, "invalid URL")
		}
		var existsErr *model.AlreadyExistsError
		if errors.As(err, &existsErr) {
			return &grpcapi.URLShortenResponse{}, status.Error(codes.AlreadyExists, "URL already exists")
		}
		return nil, status.Error(codes.Internal, "failed to shorten URL")
	}

	resp := &grpcapi.URLShortenResponse{}
	resp.SetResult(shortURL)
	return resp, nil
}

func (h *ShortenerHandler) ExpandURL(ctx context.Context, req *grpcapi.URLExpandRequest) (*grpcapi.URLExpandResponse, error) {
	originalURL, err := h.urlService.Expand(ctx, req.GetId())
	if err != nil {
		if err == model.ErrNotFound {
			return nil, status.Error(codes.NotFound, "URL not found")
		}
		if err == model.ErrLinkIsGone {
			return nil, status.Error(codes.NotFound, "URL has been deleted")
		}
		return nil, status.Error(codes.Internal, "failed to expand URL")
	}

	resp := &grpcapi.URLExpandResponse{}
	resp.SetResult(originalURL)
	return resp, nil
}

func (h *ShortenerHandler) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*grpcapi.UserURLsResponse, error) {
	userID, ok := ctx.Value(model.UserIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get user ID")
	}

	links, err := h.urlService.GetUserLinks(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get user URLs")
	}

	urlData := make([]*grpcapi.URLData, 0, len(links))
	for _, link := range links {
		if !link.IsDeleted {
			data := &grpcapi.URLData{}
			data.SetShortUrl(link.ShortURL)
			data.SetOriginalUrl(link.OriginalURL)
			urlData = append(urlData, data)
		}
	}

	resp := &grpcapi.UserURLsResponse{}
	resp.SetUrl(urlData)
	return resp, nil
}
