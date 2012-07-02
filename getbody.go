package imap

import (
	"bytes"
	"net/mail"
	"mime"
	"fmt"
	"mime/multipart"
	"encoding/base64"
	"github.com/googollee/go-encoding-ex"
	"io"
	"io/ioutil"
	"strings"
)

func GetBody(msg *mail.Message, preferType string) (string, string, string, error) {
	var mediatype string
	var typeparams map[string]string
Select:
	for {
		var err error
		mediatype, typeparams, err = mime.ParseMediaType(msg.Header.Get("Content-Type"))
		if err != nil {
			return "", "", "", err
		}
		err = nil
		switch mediatype {
		case "multipart/alternative":
			fallthrough
		case "multipart/mixed":
			fallthrough
		case "multipart/related":
			msg, err = selectMultiPart(msg, typeparams["boundary"], preferType)
			if err != nil {
				break Select
			}
		default:
			break Select
		}
		if err != nil {
			return "", "", "", fmt.Errorf("parse mail content error: %s", err)
		}
	}

	encoding := msg.Header.Get("Content-Transfer-Encoding")
	var reader io.Reader
	switch encoding {
	case "base64":
		reader = base64.NewDecoder(base64.StdEncoding, encodingex.NewIgnoreReader(msg.Body, []byte("\r\n")))
	case "quoted-printable":
		reader = encodingex.NewQuotedPrintableDecoder(msg.Body)
	default:
		reader = msg.Body
	}

	body, err := ioutil.ReadAll(reader)
	return string(body), mediatype, typeparams["charset"], err
}

func selectMultiPart(msg *mail.Message, boundary, preferType string) (*mail.Message, error) {
	reader := multipart.NewReader(msg.Body, boundary)
	parts := make(map[string]*mail.Message)
	for part, err := reader.NextPart(); err != io.EOF; part, err = reader.NextPart() {
		if err != nil {
			return nil, err
		}
		header := mail.Header(part.Header)
		mediatype, _, err := mime.ParseMediaType(header.Get("Content-Type"))
		if err != nil {
			continue
		}
		if mediatype == preferType {
			return &mail.Message{header, part}, nil
		}
		types := strings.Split(mediatype, "/")
		if len(types) == 0 {
			continue
		}
		if _, ok := parts[types[0]]; !ok {
			body, err := ioutil.ReadAll(part)
			if err == nil {
				parts[types[0]] = &mail.Message{header, bytes.NewBuffer(body)}
			}
		}
	}
	types := strings.Split(preferType, "/")
	if part, ok := parts[types[0]]; ok {
		return part, nil
	}
	if part, ok := parts["multipart"]; ok {
		return part, nil
	}
	return nil, fmt.Errorf("No prefered part")
}
