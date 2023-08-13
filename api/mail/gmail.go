package mail

import (
	"context"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type GmailClient struct {
	token       *oauth2.Token
	oauthConfig *oauth2.Config
	s           *gmail.Service
}

func NewGmailClient(
	token *oauth2.Token,
	oauthConfig *oauth2.Config,
) api.Mail {
	return &GmailClient{token: token, oauthConfig: oauthConfig}
}

func (g *GmailClient) refreshToken(ctx context.Context) error {
	newToken, err := g.oauthConfig.TokenSource(ctx, g.token).Token()
	if err != nil {
		return err
	}

	if !g.token.Valid() || g.s == nil {
		g.s, _ = gmail.NewService(
			context.Background(),
			option.WithTokenSource(oauth2.StaticTokenSource(newToken)),
			option.WithScopes(gmail.GmailReadonlyScope),
		)
	}

	g.token = newToken
	return token.Save("token.json", newToken)
}

func (g *GmailClient) GetUnreadEmails(ctx context.Context) {
	if err := g.refreshToken(ctx); err != nil {
		zap.L().Error("failed to refresh token", zap.Error(err))
		return
	}

	emails, err := g.getUnreadEmails()
	if err != nil {
		zap.L().Error("failed to get unread emails", zap.Error(err))
		return
	}

	for _, email := range emails {
		zap.L().Info("processing email", zap.Any("email", email))
	}
}

func (g *GmailClient) getUnreadEmails() ([]*gmail.Message, error) {
	lst, err := g.s.Users.Messages.List("me").Q("is:unread").MaxResults(5).Do()
	if err != nil {
		return nil, err
	}

	var emails []*gmail.Message
	for _, msg := range lst.Messages {
		email, e := g.s.Users.Messages.Get("me", msg.Id).Do()
		if e != nil {
			return nil, e
		}

		emails = append(emails, email)
	}

	return emails, nil
}
