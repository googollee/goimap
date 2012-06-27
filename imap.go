package imap

import (
	"crypto/tls"
	"fmt"
	"strings"
	"bufio"
	"bytes"
	"errors"
	"net/textproto"
	"net/mail"
)

const (
	RFC822Header = "rfc822.header"
	RFC822Text = "rfc822.text"
	Seen = "\\Seen"
	Inbox = "INBOX"
)

type IMAPClient struct {
	conn *tls.Conn
	count int
	buf []byte
}

func NewClient(addr string) (*IMAPClient, error) {
	buf := make([]byte, 1024)
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
REPLY:
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return nil, err
		}
		for _, i := range buf[:n] {
			if i == byte('\n') {
				break REPLY
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return &IMAPClient{
		conn: conn,
		buf: buf,
	}, nil
}

func (c *IMAPClient) Close() error {
	return c.conn.Close()
}

func (c *IMAPClient) Do(cmd string) *Response {
	c.count++
	cmd = fmt.Sprintf("a%03d %s\r\n", c.count, cmd)
	ret := NewResponse()

	_, err := c.conn.Write([]byte(cmd))
	if err != nil {
		ret.err = err
		return ret
	}

	for {
		n, err := c.conn.Read(c.buf)
		if err != nil {
			ret.err = err
			return ret
		}
		isFinished, err := ret.Feed(c.buf[:n])
		if err != nil {
			ret.err = err
			return ret
		}
		if isFinished {
			break
		}
	}
	return ret
}

func (c *IMAPClient) Login(user, password string) error {
	resp := c.Do(fmt.Sprintf("LOGIN %s %s", user, password))
	return resp.err
}

func (c *IMAPClient) Select(box string) *Response {
	return c.Do(fmt.Sprintf("SELECT %s", box))
}

func (c *IMAPClient) Search(flag string) ([]string, error) {
	resp := c.Do(fmt.Sprintf("SEARCH %s", flag))
	if resp.Error() != nil {
		return nil, resp.Error()
	}
	for _, reply := range resp.Replys() {
		org := reply.Origin()
		if len(org) >= 6 && strings.ToUpper(org[:6]) == "SEARCH" {
			ids := strings.Trim(org[6:], " \t\n\r")
			if ids == "" {
				return nil, nil
			}
			return strings.Split(ids, " "), nil
		}
	}
	return nil, errors.New("Invalid response")
}

func (c *IMAPClient) Fetch(id, arg string) (string, error) {
	resp := c.Do(fmt.Sprintf("FETCH %s %s", id, arg))
	if resp.Error() != nil {
		return "", resp.Error()
	}
	for _, reply := range resp.Replys() {
		org := reply.Origin()
		if len(org) < len(id) || org[:len(id)] != id {
			continue
		}
		org = org[len(id)+1:]
		if len(org) >= 5 && strings.ToUpper(org[:5]) == "FETCH" {
			body := reply.Content()
			i := strings.Index(body, "\n")
			return body[i+1:], nil
		}
	}
	return "", errors.New("Invalid response")
}

func (c *IMAPClient) StoreFlag(id, flag string) error {
	resp := c.Do(fmt.Sprintf("STORE %s FLAGS %s", id, flag))
	return resp.Error()
}

func (c *IMAPClient) Logout() error {
	resp := c.Do("LOGOUT")
	return resp.Error()
}

func (c *IMAPClient) GetMessage(id string) (*mail.Message, error) {
	headerResp := c.Do(fmt.Sprintf("FETCH %s %s", id, RFC822Header))
	if headerResp.Error() != nil {
		return nil, headerResp.Error()
	}

	replys := headerResp.Replys()

	reader := textproto.NewReader(bufio.NewReader(bytes.NewBuffer(replys[0].content)))
	header, err := reader.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	bodyResp := c.Do(fmt.Sprintf("FETCH %s %s", id, RFC822Text))
	if bodyResp.Error() != nil {
		return nil, bodyResp.Error()
	}

	return &mail.Message{
		Header: mail.Header(header),
		Body: bytes.NewBuffer(bodyResp.Replys()[0].content),
	}, nil
}
