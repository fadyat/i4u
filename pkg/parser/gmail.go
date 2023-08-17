package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
	"os"
	"regexp"
	"strings"
)

func init() {
	var lg, _ = zap.NewProduction()
	zap.ReplaceGlobals(lg)
}

func cleanText(text string) string {
	text = regexp.MustCompile(`[\n\t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(` +`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func cleanHTML(raw []byte) (string, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(string(raw)))
	if err != nil {
		return "", err
	}

	document.Find("script").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
	})

	document.Find("style").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
	})

	document.Find("link").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
	})

	document.Find("meta").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
	})

	return cleanText(document.Text()), nil
}

func main() {
	bytes, err := os.ReadFile("message.json")
	if err != nil {
		zap.L().Fatal("failed to read file", zap.Error(err))
	}

	var gmailMessage gmail.Message
	if e := json.Unmarshal(bytes, &gmailMessage); e != nil {
		zap.L().Fatal("failed to unmarshal json", zap.Error(e))
	}

	parts := gmailMessage.Payload.Parts
	for _, part := range parts {
		content, e := base64.URLEncoding.DecodeString(part.Body.Data)
		if e != nil {
			zap.L().Fatal("failed to decode base64", zap.Error(e))
		}

		switch part.MimeType {
		case "text/plain":
			zap.L().Info("text/plain", zap.String("text", cleanText(string(content))))
		case "text/html":
			text, e := cleanHTML(content)
			if e != nil {
				zap.L().Fatal("failed to clean html", zap.Error(e))
			}

			zap.L().Info("text/html", zap.String("text", text))
		}
	}

	body := gmailMessage.Payload.Body
	content, e := base64.URLEncoding.DecodeString(body.Data)
	if e != nil {
		zap.L().Fatal("failed to decode base64", zap.Error(e))
	}

	cnt, err := cleanHTML(content)
	if err != nil {
		zap.L().Fatal("failed to clean html", zap.Error(err))
	}

	zap.L().Info("text/html", zap.String("text", cnt))
}
