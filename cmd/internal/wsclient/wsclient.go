package wsclient

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/gclkaze/evamon/cmd/internal/output"
)

type Client struct {
	url    string
	token  string
	logger output.Printer
	conn   *websocket.Conn
}

func NewWSClient(url string, token string, logger output.Printer) *Client {
	return &Client{
		url:    url,
		token:  strings.TrimSpace(token),
		logger: logger,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	if c.logger != nil {
		c.logger.VerboseInfo("[ws] connecting to " + c.url)
	}

	opts := &websocket.DialOptions{HTTPHeader: http.Header{}}
	if c.token != "" {
		opts.HTTPHeader.Set("Authorization", "Bearer "+c.token)
	}

	conn, _, err := websocket.Dial(ctx, c.url, opts)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	c.conn = conn
	if c.logger != nil {
		c.logger.VerboseInfo("[ws] connected")
	}
	return nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	if c.logger != nil {
		c.logger.VerboseInfo("[ws] closing connection")
	}
	err := c.conn.Close(websocket.StatusNormalClosure, "client shutdown")
	c.conn = nil
	return err
}

func (c *Client) SendJSON(ctx context.Context, v any) error {
	if c.conn == nil {
		err := ErrNotConnected()
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	b, err := json.Marshal(v)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	if c.logger != nil {
		c.logger.VerboseInfo("[ws] -> " + string(b))
	}

	writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = c.conn.Write(writeCtx, websocket.MessageText, b)
	if err != nil && c.logger != nil {
		c.logger.Error(err)
	}
	return err
}

func (c *Client) ReadJSON(ctx context.Context, out any) error {
	if c.conn == nil {
		err := ErrNotConnected()
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	readCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_, b, err := c.conn.Read(readCtx)
	if err != nil {
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	if c.logger != nil {
		c.logger.VerboseInfo("[ws] <- " + string(b))
	}

	if err := json.Unmarshal(b, out); err != nil {
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}
	return nil
}
func (c *Client) ReadJSONForever(ctx context.Context, out any) error {
	if c.conn == nil {
		err := ErrNotConnected()
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	_, b, err := c.conn.Read(ctx) // no timeout
	if err != nil {
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}

	if c.logger != nil {
		c.logger.VerboseInfo("[ws] <- " + string(b))
	}

	if err := json.Unmarshal(b, out); err != nil {
		if c.logger != nil {
			c.logger.Error(err)
		}
		return err
	}
	return nil
}
