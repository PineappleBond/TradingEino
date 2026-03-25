package svc

import (
	"context"
	"fmt"
	"os"

	"github.com/PineappleBond/TradingEino/backend/internal/config"
	"github.com/cloudwego/eino-ext/components/model/openai"
)

func mustInitChatModel(cfg config.Config) *openai.ChatModel {
	model, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:  cfg.ChatModel.APIKey,
		BaseURL: cfg.ChatModel.BaseURL,
		Model:   cfg.ChatModel.Model,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init chat model: %v\n", err)
		os.Exit(1)
	}
	return model
}
