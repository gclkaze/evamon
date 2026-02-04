package auth

import (
	"fmt"
	"os"
	"strings"

	"github.com/magiconair/properties"
)

type TokenLoader struct {
	props        *properties.Properties
	tokenFileKey string // e.g. "auth.token_file"
}

func NewTokenLoader(props *properties.Properties, tokenFileKey string) *TokenLoader {
	return &TokenLoader{
		props:        props,
		tokenFileKey: tokenFileKey,
	}
}

func (l *TokenLoader) LoadToken() (string, error) {
	if l.props == nil {
		return "", fmt.Errorf("properties is nil")
	}

	tokenFilePath := strings.TrimSpace(l.props.GetString(l.tokenFileKey, ""))
	if tokenFilePath == "" {
		return "", fmt.Errorf("token file path is empty (application.properties key %q)", l.tokenFileKey)
	}

	b, err := os.ReadFile(tokenFilePath)
	if err != nil {
		return "", fmt.Errorf("read token file %q: %w", tokenFilePath, err)
	}

	token := strings.TrimSpace(string(b))
	if token == "" {
		return "", fmt.Errorf("token file %q is empty", tokenFilePath)
	}
	return token, nil
}
