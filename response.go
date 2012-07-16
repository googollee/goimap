package imap

import (
	"errors"
	"strconv"
	"strings"
)

type reply struct {
	origin  []byte
	type_   []byte
	length  []byte
	content []byte
}

func newReply() (ret reply) {
	return reply{
		origin:  make([]byte, 0, 0),
		type_:   make([]byte, 0, 0),
		content: make([]byte, 0, 0),
	}
}

func (r reply) Origin() string {
	return string(r.origin)
}

func (r reply) Type() string {
	return string(r.type_)
}

func (r reply) Length() (i int, err error) {
	i, err = strconv.Atoi(string(r.length))
	return
}

func (r reply) Content() string {
	return string(r.content)
}

type feedStatus int

const (
	feedInit feedStatus = iota
	feedStar
	feedReply
	feedReplyType
	feedReplyLength
	feedReplyContent
	feedReplyMeet0d
	feedStatusLine
	feedStatusLineMeet0d
	feedFinished
)

type Response struct {
	id     string
	status string
	err    error
	replys []reply

	buf              []byte
	feedStatus       feedStatus
	parenthesisCount int
	reply            reply
}

func NewResponse() *Response {
	return &Response{
		buf:              make([]byte, 0, 0),
		feedStatus:       feedInit,
		parenthesisCount: 0,
	}
}

func (r *Response) Feed(input []byte) (bool, error) {
	for _, i := range input {
		switch r.feedStatus {
		case feedInit:
			if i == byte('*') {
				r.feedStatus = feedStar
			} else {
				r.feedStatus = feedStatusLine
				r.buf = append(r.buf, i)
			}
		case feedStar:
			if i != byte(' ') {
				r.feedStatus = feedReply
				r.reply = newReply()
				r.reply.origin = append(r.reply.origin, i)
			}
		case feedReply:
			switch i {
			case byte('\r'):
				r.feedStatus = feedReplyMeet0d
			case byte('('):
				r.feedStatus = feedReplyType
				r.reply.origin = append(r.reply.origin, i)
			default:
				r.reply.origin = append(r.reply.origin, i)
			}
		case feedReplyType:
			switch i {
			case byte(')'):
				r.feedStatus = feedReply
			case byte(' '):
				if len(r.reply.type_) > 0 {
					r.feedStatus = feedReplyLength
				}
			default:
				r.reply.type_ = append(r.reply.type_, i)
			}
			r.reply.origin = append(r.reply.origin, i)
		case feedReplyLength:
			r.reply.origin = append(r.reply.origin, i)
			if i == byte('\n') {
				r.feedStatus = feedReplyContent
			}
			if byte('0') <= i && i <= byte('9') {
				r.reply.length = append(r.reply.length, i)
			}
		case feedReplyContent:
			r.reply.origin = append(r.reply.origin, i)
			r.reply.content = append(r.reply.content, i)
			i, err := r.reply.Length()
			if err != nil {
				return false, errors.New("Parse response error, reply need a valid length number")
			}
			if len(r.reply.content) == i {
				r.feedStatus = feedReply
			}
		case feedReplyMeet0d:
			if i == byte('\n') {
				r.feedStatus = feedInit
				r.replys = append(r.replys, r.reply)
				r.buf = r.buf[0:0]
			} else {
				r.feedStatus = feedReply
				r.reply.origin = append(r.reply.origin, i)
			}
		case feedStatusLine:
			if i == byte('\r') {
				r.feedStatus = feedStatusLineMeet0d
			} else {
				r.buf = append(r.buf, i)
			}
		case feedStatusLineMeet0d:
			if i == byte('\n') {
				r.feedStatus = feedFinished
				array := strings.SplitN(string(r.buf), " ", 2)
				if len(array) > 0 {
					r.id = array[0]
				}
				if len(array) > 1 {
					r.status = array[1]
				}
				if len(r.status) < 3 || r.status[:3] != "OK " {
					r.err = errors.New(r.status)
				}
				return true, nil
			} else {
				r.feedStatus = feedStatusLine
				r.buf = append(r.buf, byte('\r'), i)
			}
		case feedFinished:
			return true, errors.New("Need no more feed")
		}
	}
	return false, nil
}

func (r *Response) Id() string {
	return r.id
}

func (r *Response) Status() string {
	return r.status
}

func (r *Response) Error() error {
	return r.err
}

func (r *Response) Replys() []reply {
	return r.replys
}
