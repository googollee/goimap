package imap

import (
	"testing"
)

func TestResponse(t *testing.T) {
	line1 := "* 6955 FETCH (RFC822.HEADER {499}\r\nMIME-Version: 1.0\r"
	line2 := "\nReceived: by 10.76.101.172 with HTTP; Tue, 26 Jun 2012 23:11:28 -0700 (PDT)\r\n"
	line3 := "Date: Wed, 27 Jun 2012 14:11:28 +0800\r\nDelivered-To: googollee@gmail.com\r\nMessage-ID: "
	line4 := "<CAOf82vP-CNcxcNKvRSHc_rGrNrEoTq7DLOEckuD1g-MN7LqtVg@mail.gmail.com>\r\nSubject: test\r\nFrom: Googol Lee <googollee@gmail.com>\r\n"
	line5 := "To: =?UTF-8?B?R29vZ29sIExlZSAtIEdvb2dsZee6r+eIt+S7rO+8gemTgeihgOecn+axieWtkO+8ge+8gQ==?=\r\n <googollee@gmail.com>"
	line6 := "\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: base64\r\n\r\n)\r\n"
	line7 := "a007 OK Success\r\n"
	line8 := "fdafas"

	lines := []string{line1, line2, line3, line4, line5, line6, line7, line8}
	resp := NewResponse()
	for _, line := range lines {
		isFinished, _ := resp.Feed([]byte(line))
		if isFinished {
			break
		}
	}

	if resp.Id() != "a007" {
		t.Errorf("resp.Id should a007, got: %s", resp.Id())
	}
	if resp.Status() != "OK Success" {
		t.Errorf("resp.Status should OK, got: %s", resp.Status())
	}
	if resp.Error() != nil {
		t.Errorf("resp.Error should nil, got: %s", resp.Error())
	}
	if len(resp.Replys()) != 1 {
		t.Errorf("resp.Reply should len 1, got: %v", resp.Replys())
	}
	replys := resp.Replys()
	if replys[0].Origin()[:4] != "6955" {
		t.Errorf("resp.Reply[0].Origin should start with 6909, got: %s", replys[0].Origin())
	}
	if replys[0].Type() != "RFC822.HEADER" {
		t.Errorf("resp.Reply[0].Type should be RFC822.HEADER, got: %s", replys[0].Type())
	}
	if len(replys[0].Content()) != 499 {
		t.Errorf("resp.Reply[0].Content length should be 499, got: %d", len(replys[0].Content()))
	}
	if replys[0].Content()[:4] != "MIME" {
		t.Errorf("resp.Reply[0].InParenthesis should start with MIME, got: %s", replys[0].Content())
	}
}
