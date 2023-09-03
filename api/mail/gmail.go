package mail

import (
	"context"
	"fmt"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/config"
	"github.com/fadyat/i4u/internal/entity"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailClient struct {
	cfg         *config.Gmail
	token       *oauth2.Token
	oauthConfig *oauth2.Config
	s           *gmail.Service
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

func (g *GmailClient) CreateLabel(ctx context.Context, label string) (*gmail.Label, error) {
	if err := g.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %w", err)
	}

	return g.s.Users.Labels.Create("me", &gmail.Label{Name: label}).Do()
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
	return token.Save(g.cfg.TokenFile, newToken)
}

func (g *GmailClient) GetUnreadMsgs(ctx context.Context) ([]*gmail.Message, error) {
	if err := g.refreshToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to refresh access token: %s", err)
	}

	msgs, err := g.getUnreadMsgs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get unread messages: %s", err)
	}

	return msgs, nil
}

func (g *GmailClient) getUnreadMsgs(ctx context.Context) ([]*gmail.Message, error) {
	unread, err := g.s.Users.Messages.List("me").
		Q("in:inbox -label:i4u").
		MaxResults(1).
		Context(ctx).
		Do()

	if err != nil {
		return nil, err
	}

	var msgs = make([]*gmail.Message, len(unread.Messages))
	for i, msg := range unread.Messages {
		// todo: do in parallel
		msgs[i], err = g.s.Users.Messages.Get("me", msg.Id).
			Format("full").
			Context(ctx).
			Do()

		if err != nil {
			return nil, err
		}
	}

	return msgs, nil
}
