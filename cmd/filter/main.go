package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/segmentio/kafka-go"

	"kafka-bridge/internal/config"
	kafkapkg "kafka-bridge/internal/kafka"
)

func main() {
	var cfgPath string
	flag.StringVar(&cfgPath, "config", "config/config.yaml", "path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	writerPool := kafkapkg.NewWriterPool(cfg.Brokers, cfg.ClientID)
	defer func() {
		if err := writerPool.Close(); err != nil {
			log.Printf("close writers: %v", err)
		}
	}()

	var wg sync.WaitGroup
	for _, route := range cfg.Routes {
		route := route
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := streamRoute(ctx, cfg, route, writerPool); err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("route %s stopped: %v", route.DisplayName(), err)
			}
		}()
	}

	wg.Wait()
}

func streamRoute(ctx context.Context, cfg *config.Config, route config.Route, writers *kafkapkg.WriterPool) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        fmt.Sprintf("%s-%s", cfg.GroupID, slug(route.DisplayName())),
		GroupTopics:    route.SourceTopics,
		CommitInterval: cfg.CommitInterval,
	})
	defer reader.Close()

	log.Printf("route %s listening to %s -> %s", route.DisplayName(), strings.Join(route.SourceTopics, ","), route.DestinationTopic)
	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			return err
		}

		match, isnValue, err := matchesISN(m.Value, route.MatchValues)
		if err != nil {
			log.Printf("route %s: skip invalid JSON: %v", route.DisplayName(), err)
			continue
		}
		if !match {
			continue
		}

		writer := writers.Get(route.DestinationTopic)
		if err := writer.WriteMessages(ctx, cloneMessage(m)); err != nil {
			log.Printf("route %s: write failed: %v", route.DisplayName(), err)
			continue
		}
		log.Printf("route %s forwarded message offset %d (isn=%s) to %s", route.DisplayName(), m.Offset, isnValue, route.DestinationTopic)
	}
}

func matchesISN(value []byte, allowed []string) (bool, string, error) {
	if len(allowed) == 0 {
		return true, "", nil
	}

	var payload map[string]any
	if err := json.Unmarshal(value, &payload); err != nil {
		return false, "", err
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		return false, "", errors.New("missing data object")
	}

	raw, ok := data["isn"]
	if !ok {
		return false, "", errors.New("data.isn not found")
	}

	isn := fmt.Sprintf("%v", raw)
	for _, candidate := range allowed {
		if candidate == isn {
			return true, isn, nil
		}
	}
	return false, isn, nil
}

func slug(in string) string {
	replacer := strings.NewReplacer(" ", "-", "/", "-", "\\", "-", ".", "-")
	return replacer.Replace(strings.ToLower(in))
}

func cloneMessage(m kafka.Message) kafka.Message {
	cloned := kafka.Message{
		Key:     append([]byte(nil), m.Key...),
		Value:   append([]byte(nil), m.Value...),
		Headers: make([]kafka.Header, len(m.Headers)),
		Time:    m.Time,
	}
	copy(cloned.Headers, m.Headers)
	return cloned
}
