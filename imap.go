package imap

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/googollee/go-encoding-ex"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
)

const (
	RFC822Header = "rfc822.header"
	RFC822Text   = "rfc822.text"
	Seen         = "\\Seen"
	Deleted      = "\\Deleted"
	Inbox        = "INBOX"
)

type IMAPClient struct {
	conn  *tls.Conn
	count int
	buf   []byte
}

func NewClient(conn net.Conn, hostname string) (*IMAPClient, error) {
	config := tls.Config{
		ServerName: hostname,
	}
	c := tls.Client(conn, &config)
	buf := make([]byte, 1024)
REPLY:
	for {
		n, err := c.Read(buf)
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
		conn: c,
		buf:  buf,
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
		Body:   bytes.NewBuffer(bodyResp.Replys()[0].content),
	}, nil
}

func ParseAddress(str string) ([]*mail.Address, error) {
	inQuote := false
	lastStart := 0
	strs := make([]string, 0, 0)
	for i, c := range str {
		switch c {
		case '"':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				strs = append(strs, str[lastStart:i])
				lastStart = i + 1
			}
		}
	}
	strs = append(strs, str[lastStart:len(str)])
	ret := make([]*mail.Address, len(strs), len(strs))
	for i, s := range strs {
		if s[len(s)-1] == '>' {
			split := strings.LastIndex(s, "<")
			name := strings.Trim(s[:split], "\" ")
			addr := s[split:]
			if name[0] == '=' {
				data, charset, err := encodingex.DecodeEncodedWord(name)
				if err != nil {
					return nil, fmt.Errorf("address %d invalid: %s", i, err)
				}
				data, err = encodingex.Conv(data, "UTF-8", charset)
				if err != nil {
					return nil, fmt.Errorf("address %d convert charset error: %s", i, err)
				}
				ret[i] = &mail.Address{
					Name:    data,
					Address: strings.Trim(addr, "<>"),
				}
			} else {
				ret[i] = &mail.Address{
					Name:    strings.Trim(name, "\""),
					Address: strings.Trim(addr, "<>"),
				}
			}
		} else {
			ret[i] = &mail.Address{
				Address: s,
			}
		}
	}
	return ret, nil
}
