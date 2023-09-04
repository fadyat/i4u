package mail

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"github.com/fadyat/i4u/pkg/syncs"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"sync"
	"time"
)

type GmailClient struct {
	cfg         *config.Gmail
	token       *oauth2.Token
	oauthConfig *oauth2.Config
	s           *gmail.Service

	// tknMtx is used to prevent concurrent access to the token file
	// when it is being refreshed.
	tknMtx sync.Mutex
}

func NewGmailClient(
	tkn *oauth2.Token,
	oauthConfig *oauth2.Config,
	gmailConfig *config.Gmail,
) api.Mail {
	return &GmailClient{
		token:       tkn,
		oauthConfig: oauthConfig,
		cfg:         gmailConfig,
		tknMtx:      sync.Mutex{},
	}
}

func (g *GmailClient) LabelMsg(ctx context.Context, msg entity.MessageForLabeler) error {
	if err := g.refreshToken(ctx); err != nil {
		return fmt.Errorf("failed to refresh access token: %w", err)
	}

	_, err := g.s.Users.Messages.Modify(
		"me",
		msg.ID(),
		&gmail.ModifyMessageRequest{
			AddLabelIds: []string{msg.Label()},
		},
	).Context(ctx).Do()

	return err
}

func (g *GmailClient) CreateLabel(
	ctx context.Context, labelName string,
) (*entity.Label, error) {
	if err := g.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	l, err := g.s.Users.Labels.Create("me", &gmail.Label{Name: labelName}).
		Context(ctx).
		Do()

	if err != nil {
		return nil, err
	}

	return entity.NewLabelFromGmailLabel(l), nil
}

func (g *GmailClient) refreshToken(ctx context.Context) error {
	newToken, err := g.oauthConfig.TokenSource(ctx, g.token).Token()
	if err != nil {
		return err
	}

	if !g.token.Valid() || g.s == nil {
		g.s, _ = gmail.NewService(
			context.Background(),
			option.WithTokenSource(
				oauth2.StaticTokenSource(newToken),
			),
			option.WithScopes(
				gmail.GmailReadonlyScope,
				gmail.GmailLabelsScope,
				gmail.GmailModifyScope,
			),
		)
	}
	g.token = newToken

	g.tknMtx.Lock()
	defer g.tknMtx.Unlock()

	return token.Save(g.cfg.TokenFile, newToken)
}

func (g *GmailClient) GetUnreadMsgs(ctx context.Context) <-chan entity.MessageWithError {
	wrappedMsgsCh := make(chan entity.MessageWithError)

	go func() {
		defer close(wrappedMsgsCh)

		if err := g.refreshToken(ctx); err != nil {
			wrappedMsgsCh <- entity.MessageWithError{
				Err: fmt.Errorf("failed to refresh access token: %w", err),
			}
			return
		}

		if err := g.getUnreadMsgs(ctx, wrappedMsgsCh); err != nil {
			wrappedMsgsCh <- entity.MessageWithError{
				Err: fmt.Errorf("failed to get unread messages: %w", err),
			}
			return
		}
	}()

	return wrappedMsgsCh
}

// getFullMessageContent will get the full message content and
// parse it to an entity.Message, then it will push it to the
// wrappedMsgsCh channel to be consumed by the next jobs in the
// pipeline.
func (g *GmailClient) getFullMessageContent(
	ctx context.Context,
	id string,
	wrappedMsgsCh chan<- entity.MessageWithError,
) {
	msg, err := g.s.Users.Messages.Get("me", id).
		Format("full").
		Context(ctx).
		Do()

	if err != nil {
		wrappedMsgsCh <- entity.MessageWithError{
			Err: fmt.Errorf("failed to get message: %w", err),
		}
		return
	}

	parsed, e := entity.NewMsgFromGmailMessage(msg)
	if e != nil {
		wrappedMsgsCh <- entity.MessageWithError{
			Err: fmt.Errorf("failed to parse message: %w", e),
		}
		return
	}

	zap.S().Debugf("got message: %s", parsed.ID())
	wrappedMsgsCh <- entity.MessageWithError{
		Msg: parsed.WithLabel(g.cfg.L.I4U),
	}
}

// getUnreadMsgs will get all basic info about the unread messages
// and launch a goroutine for each message to get the full message
// content.
func (g *GmailClient) getUnreadMsgs(
	ctx context.Context,
	wrappedMsgsCh chan<- entity.MessageWithError,
) error {
	unread, err := g.s.Users.Messages.List("me").
		Q("in:inbox -label:i4u").
		MaxResults(g.cfg.MessagesLimit).
		Context(ctx).
		Do()

	if err != nil {
		return err
	}

	var wg syncs.WaitGroup
	for _, msg := range unread.Messages {
		id := msg.Id

		wg.Go(func() {
			timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			g.getFullMessageContent(timeout, id, wrappedMsgsCh)
		})
	}

	wg.Wait()
	return nil
}
