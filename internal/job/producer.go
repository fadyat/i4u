package job

import (
	"context"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
	"time"
)

type producer struct {
	mailClient     api.Mail
	analyzerClient api.Analyzer
	summarizer     api.Summarizer
	sender         api.Sender
}

func NewProducer(
	mailClient api.Mail,
	analyzerClient api.Analyzer,
	summarizer api.Summarizer,
	sender api.Sender,
) Producer {
	return &producer{
		mailClient:     mailClient,
		analyzerClient: analyzerClient,
		summarizer:     summarizer,
		sender:         sender,
	}
}

func (p *producer) Produce(ctx context.Context) <-chan error {
	errsCh := make(chan error)

	labelerChan := make(chan entity.Message)
	analyzerChan := make(chan entity.Message)
	summarizerChan := make(chan entity.Message)
	senderChan := make(chan entity.SummaryMsg)

	fetcherJob := NewFetcherJob(
		p.mailClient, errsCh, []chan<- entity.Message{labelerChan, analyzerChan},
	)
	labelerJob := NewLabelerJob(
		p.mailClient, errsCh, labelerChan,
	)
	analyzerJob := NewAnalyzerJob(
		p.analyzerClient, errsCh, analyzerChan, []chan<- entity.Message{summarizerChan, labelerChan},
	)
	summarizerJob := NewSummarizerJob(
		p.summarizer, errsCh, summarizerChan, senderChan,
	)
	senderJob := NewSenderJob(
		p.sender, errsCh, senderChan,
	)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		zap.S().Info("starting fetcher job")
		for {
			select {
			case <-ticker.C:
				zap.S().Debug("fetching messages")
				fetcherJob.Run(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()

	// todo: write in more compact way
	//  like struct with channel and job
	go func() {
		defer close(labelerChan)

		zap.S().Info("starting labeler job")
		labelerJob.Run(ctx)
	}()

	go func() {
		defer close(analyzerChan)

		zap.S().Info("starting analyzer job")
		analyzerJob.Run(ctx)
	}()

	go func() {
		defer close(summarizerChan)

		zap.S().Info("starting summarizer job")
		summarizerJob.Run(ctx)
	}()

	go func() {
		defer close(senderChan)

		zap.S().Info("starting sender job")
		senderJob.Run(ctx)
	}()

	go func() {
		defer close(errsCh)
		<-ctx.Done()
	}()

	return errsCh
}
