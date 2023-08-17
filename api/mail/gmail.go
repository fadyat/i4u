package mail

import (
	"context"
	"github.com/fadyat/i4u/api"
	"github.com/fadyat/i4u/cmd/i4u/token"
	"github.com/fadyat/i4u/internal/entity"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"os"
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

func (g *GmailClient) GetUnreadMsgs(ctx context.Context) {
	if err := g.refreshToken(ctx); err != nil {
		zap.L().Error("failed to refresh token", zap.Error(err))
		return
	}

	msgs, err := g.getUnreadMsgs()
	if err != nil {
		zap.L().Error("failed to get unread messages", zap.Error(err))
		return
	}

	for _, msg := range msgs {
		customMsg, e := entity.NewMsgFromGmailMessage(msg)
		if e != nil {
			zap.L().Error("failed to create custom message", zap.Error(e))
			return
		}

		zap.L().Info("new unread message", zap.Any("message", customMsg))
	}
}

func (g *GmailClient) getUnreadMsgs() ([]*gmail.Message, error) {
	unread, err := g.s.Users.Messages.List("me").MaxResults(5).Do()
	if err != nil {
		return nil, err
	}

	var msgs = make([]*gmail.Message, len(unread.Messages))
	for i, msg := range unread.Messages {
		msgs[i], err = g.s.Users.Messages.Get("me", msg.Id).Format("full").Do()
		if err != nil {
			return nil, err
		}

		file, err := os.Create("message.json")
		if err != nil {
			return nil, err
		}

		j, _ := msgs[i].MarshalJSON()
		_, err = file.WriteString(string(j))
		if err != nil {
			return nil, err
		}
	}

	return msgs, nil
}
