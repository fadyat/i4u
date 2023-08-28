package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"google.golang.org/api/gmail/v1"
	"os"
	"regexp"
	"strings"
)

const (
	PlainText = "text/plain"
	HTMLText  = "text/html"
)

// CleanMsg getting message body or combining all parts of message body and
// clean all unnecessary data from text/plain or text/html formats to return
// it in valid format for analyzing.
func CleanMsg(msg *gmail.Message, takeFormat string) (string, error) {
	body := msg.Payload.Body
	if body.Data != "" {
		return cleanBody(body, msg.Payload.MimeType)
	}

	parts := msg.Payload.Parts
	if len(parts) == 0 {
		return "", errors.New("no body or parts found")
	}

	return cleanParts(parts, takeFormat)
}

func cleanParts(parts []*gmail.MessagePart, takeFormat string) (string, error) {
	var content strings.Builder
	for _, part := range parts {
		if part.MimeType != takeFormat {
			continue
		}

		p, err := cleanBody(part.Body, takeFormat)
		if err != nil {
			return "", err
		}

		content.WriteString(p)
	}

	return content.String(), nil
}

func cleanBody(body *gmail.MessagePartBody, bodyFormat string) (string, error) {
	content, err := base64.URLEncoding.DecodeString(body.Data)
	if err != nil {
		return "", err
	}

	return cleanContent(string(content), bodyFormat)
}

func cleanContent(content, bodyFormat string) (string, error) {
	switch bodyFormat {
	case PlainText:
		return cleanText(content), nil
	case HTMLText:
		return cleanHTML(content)
	}

	return "", fmt.Errorf("unknown body format: %s", bodyFormat)
}

// cleanText removes all unnecessary data from text/plain
// and returns only text content.
func cleanText(text string) string {
	pipe := []func(string) string{
		replaceNoneBreakingSpace,
		removeUrls,
		removeSpaces,
	}

	for _, fn := range pipe {
		text = fn(text)
	}

	return text
}

func replaceNoneBreakingSpace(text string) string {
	return strings.ReplaceAll(text, "\u00a0", " ")
}

func removeUrls(text string) string {
	return regexp.MustCompile(`https?://\S+`).ReplaceAllString(text, "")
}

func removeSpaces(text string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
}

// cleanHTML removes all unnecessary data from html text
// and returns only text content.
func cleanHTML(raw string) (string, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(raw))
	if err != nil {
		return "", err
	}

	removeElements := []string{"script", "style", "link", "meta"}
	for _, elem := range removeElements {
		document.Find(elem).Each(func(i int, selection *goquery.Selection) {
			selection.Remove()
		})
	}

	return cleanText(document.Text()), nil
}

func main() {
	var lg, _ = zap.NewProduction()
	zap.ReplaceGlobals(lg)

	bytes, err := os.ReadFile("message.json")
	if err != nil {
		zap.L().Fatal("failed to read file", zap.Error(err))
	}

	var gmailMessage gmail.Message
	if e := json.Unmarshal(bytes, &gmailMessage); e != nil {
		zap.L().Fatal("failed to unmarshal json", zap.Error(e))
	}

	msg, err := CleanMsg(&gmailMessage, PlainText)
	if err != nil {
		zap.L().Fatal("failed to clean message", zap.Error(err))
	}

	cleanedFile, err := os.Create("cleaned.txt")
	if err != nil {
		zap.L().Fatal("failed to create file", zap.Error(err))
	}

	_, err = cleanedFile.WriteString(msg)
	if err != nil {
		zap.L().Fatal("failed to write to file", zap.Error(err))
	}

	zap.L().Info("cleaned message saved to file")
}
