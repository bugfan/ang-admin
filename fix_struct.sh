#!/bin/bash
sed -i '' 's/type EmbedClient struct {/type EmbedClient struct {\n\tclient *openai.Client\n\tmodel  string\n\tisGemini bool\n\tgeminiKey string\n}\n\ntype EmbedClientOld struct {/g' internal/rag/rag.go
