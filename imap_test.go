package imap

import (
	"bufio"
	"bytes"
	"net/mail"
	"net/textproto"
	"testing"
)

func TestResponse(t *testing.T) {
	line1 := "* 6955 FETCH (RFC822.HEADER {499}\r\nMIME-Version: 1.0\r"
	line2 := "\nReceived: by 10.76.101.172 with HTTP; Tue, 26 Jun 2012 23:11:28 -0700 (PDT)\r\n"
	line3 := "Date: Wed, 27 Jun 2012 14:11:28 +0800\r\nDelivered-To: googollee@gmail.com\r\nMessage-ID: "
	line4 := "<CAOf82vP-CNcxcNKvRSHc_rGrNrEoTq7DLOEckuD1g-MN7LqtVg@mail.gmail.com>\r\nSubject: test\r\nFrom: Googol Lee <googollee@gmail.com>\r\n"
	line5 := "To: =?UTF-8?B?R29vZ29sIExlZSAtIEdvb2dsZee6r+eIt+S7rO+8gemTgeihgOecn+axieWtkO+8ge+8gQ==?=\r\n <googollee@gmail.com>"
	line6 := "\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: base64\r\n\r\n"
	linea := "FLAG (\\Seen))\r\n"
	line7 := "a007 OK Success\r\n"
	line8 := "fdafas"

	lines := []string{line1, line2, line3, line4, line5, line6, linea, line7, line8}
	resp := NewResponse()
	isFinished := false
	for _, line := range lines {
		isFinished, _ = resp.Feed([]byte(line))
		if isFinished {
			break
		}
	}

	if !isFinished {
		t.Errorf("should finished")
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
	if replys[0].Type() != "RFC822.HEADER\\Seen" {
		t.Errorf("resp.Reply[0].Type should be RFC822.HEADER\\Seen, got: %s", replys[0].Type())
	}
	if len(replys[0].Content()) != 499 {
		t.Errorf("resp.Reply[0].Content length should be 499, got: %d", len(replys[0].Content()))
	}
	if replys[0].Content()[:4] != "MIME" {
		t.Errorf("resp.Reply[0].InParenthesis should start with MIME, got: %s", replys[0].Content())
	}
}

func TestParseAddress(t *testing.T) {
	{
		addrs, _ := ParseAddress("=?GB2312?B?1arSqsrVvP7Iyw==?= <pongba@googlegroups.com>")
		if len(addrs) != 1 {
			t.Errorf("expect: 1, got: %d", len(addrs))
		}
		if addrs[0].Name != "摘要收件人" {
			t.Errorf("expect: 摘要收件人, got: %s", addrs[0].Name)
		}
		if addrs[0].Address != "pongba@googlegroups.com" {
			t.Errorf("expect: pongba@googlegroups.com, got: %s", addrs[0].Address)
		}
	}

	{
		addrs, _ := ParseAddress("\"abc, 123\" <pongba@googlegroups.com>")
		if len(addrs) != 1 {
			t.Errorf("expect: 1, got: %d", len(addrs))
		}
		if addrs[0].Name != "abc, 123" {
			t.Errorf("expect: abc, 123, got: %s", addrs[0].Name)
		}
		if addrs[0].Address != "pongba@googlegroups.com" {
			t.Errorf("expect: pongba@googlegroups.com, got: %s", addrs[0].Address)
		}
	}

	{
		addrs, _ := ParseAddress("pongba@googlegroups.com")
		if len(addrs) != 1 {
			t.Errorf("expect: 1, got: %d", len(addrs))
		}
		if addrs[0].Name != "" {
			t.Errorf("expect: empty string, got: %s", addrs[0].Name)
		}
		if addrs[0].Address != "pongba@googlegroups.com" {
			t.Errorf("expect: pongba@googlegroups.com, got: %s", addrs[0].Address)
		}
	}

	{
		addrs, _ := ParseAddress("Pongba <pongba@googlegroups.com>")
		if len(addrs) != 1 {
			t.Errorf("expect: 1, got: %d", len(addrs))
		}
		if addrs[0].Name != "Pongba" {
			t.Errorf("expect: Pongba, got: %s", addrs[0].Name)
		}
		if addrs[0].Address != "pongba@googlegroups.com" {
			t.Errorf("expect: pongba@googlegroups.com, got: %s", addrs[0].Address)
		}
	}

	{
		addrs, _ := ParseAddress("Pongba <pongba@googlegroups.com>, =?GB2312?B?1arSqsrVvP7Iyw==?= <pongba@googlegroups.com>")
		if len(addrs) != 2 {
			t.Errorf("expect: 1, got: %d", len(addrs))
		}
		if addrs[0].Name != "Pongba" {
			t.Errorf("expect: Pongba, got: %s", addrs[0].Name)
		}
		if addrs[0].Address != "pongba@googlegroups.com" {
			t.Errorf("expect: pongba@googlegroups.com, got: %s", addrs[0].Address)
		}
		if addrs[1].Name != "摘要收件人" {
			t.Errorf("expect: 摘要收件人, got: %s", addrs[1].Name)
		}
		if addrs[1].Address != "pongba@googlegroups.com" {
			t.Errorf("expect: pongba@googlegroups.com, got: %s", addrs[1].Address)
		}
	}
}

func TestSelectPart(t *testing.T) {
	{
		hs := `MIME-Version: 1.0
Received: by 10.76.101.172 with HTTP; Tue, 26 Jun 2012 23:11:28 -0700 (PDT)
Date: Wed, 27 Jun 2012 14:11:28 +0800
Delivered-To: googollee@gmail.com
Message-ID: <CAOf82vP-CNcxcNKvRSHc_rGrNrEoTq7DLOEckuD1g-MN7LqtVg@mail.gmail.com>
Subject: test
From: Googol Lee <googollee@gmail.com>
To: =?UTF-8?B?R29vZ29sIExlZSAtIEdvb2dsZee6r+eIt+S7rO+8gemTgeihgOecn+axieWtkO+8ge+8gQ==?= <googollee@gmail.com>
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64`
		bd := `YTAwNyBPSyBTdWNjZXNzCgopCgotLQrmlrDnmoTnkIborrrku47lsJHmlbDkurrnmoTkuLvlvKDl
iLDkuIDnu5/lpKnkuIvvvIzlubbkuI3mmK/lm6DkuLrov5nkuKrnkIborrror7TmnI3kuobliKvk
urrmipvlvIPml6fop4LngrnvvIzogIzmmK/lm6DkuLrkuIDku6PkurrnmoTpgJ3ljrvjgIIK`
		reader := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(hs)))
		hr, _ := reader.ReadMIMEHeader()
		msg := &mail.Message{
			Header: mail.Header(hr),
			Body:   bytes.NewBufferString(bd),
		}

		expect := "a007 OK Success\n\n)\n\n--\n新的理论从少数人的主张到一统天下，并不是因为这个理论说服了别人抛弃旧观点，而是因为一代人的逝去。\n"
		content, mediatype, charset, _ := GetBody(msg, "text/plain")
		if mediatype != "text/plain" {
			t.Errorf("mediatype expect: text/plain, got: %s", mediatype)
		}
		if charset != "UTF-8" {
			t.Errorf("charset expect: UTF-8, got: %s", charset)
		}
		if content != expect {
			t.Errorf("content expect:\n%s\ngot:\n%s", expect, content)
		}
	}

	{
		hs := `MIME-Version: 1.0
Received: by 10.76.101.172 with HTTP; Tue, 26 Jun 2012 23:11:28 -0700 (PDT)
Date: Wed, 27 Jun 2012 14:11:28 +0800
Delivered-To: googollee@gmail.com
Message-ID: <CAOf82vP-CNcxcNKvRSHc_rGrNrEoTq7DLOEckuD1g-MN7LqtVg@mail.gmail.com>
Subject: test
From: Googol Lee <googollee@gmail.com>
To: =?UTF-8?B?R29vZ29sIExlZSAtIEdvb2dsZee6r+eIt+S7rO+8gemTgeihgOecn+axieWtkO+8ge+8gQ==?= <googollee@gmail.com>
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: quoted-printable`
		bd := "If you believe that truth=3Dbeauty, then surely =\r\nmathematics is the most beautiful branch of philosophy."
		reader := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(hs)))
		hr, _ := reader.ReadMIMEHeader()
		msg := &mail.Message{
			Header: mail.Header(hr),
			Body:   bytes.NewBufferString(bd),
		}

		expect := "If you believe that truth=beauty, then surely mathematics is the most beautiful branch of philosophy."
		content, mediatype, charset, _ := GetBody(msg, "text/plain")
		if mediatype != "text/plain" {
			t.Errorf("mediatype expect: text/plain, got: %s", mediatype)
		}
		if charset != "UTF-8" {
			t.Errorf("charset expect: UTF-8, got: %s", charset)
		}
		if content != expect {
			t.Errorf("content expect:\n%s\ngot:\n%s", expect, content)
		}
	}

	{
		hs := `MIME-Version: 1.0
Received: by 10.76.101.172 with HTTP; Tue, 26 Jun 2012 23:11:28 -0700 (PDT)
Date: Wed, 27 Jun 2012 14:11:28 +0800
Delivered-To: googollee@gmail.com
Message-ID: <CAOf82vP-CNcxcNKvRSHc_rGrNrEoTq7DLOEckuD1g-MN7LqtVg@mail.gmail.com>
Subject: test
From: Googol Lee <googollee@gmail.com>
To: =?UTF-8?B?R29vZ29sIExlZSAtIEdvb2dsZee6r+eIt+S7rO+8gemTgeihgOecn+axieWtkO+8ge+8gQ==?= <googollee@gmail.com>
Content-Type: multipart/alternative; boundary=14dae9d67df4436caa04c39ae8b8`
		bd := `--14dae9d67df4436caa04c39ae8b8
Content-Type: text/plain; charset=UTF-8
Content-Transfer-Encoding: base64

YWJjYWJjCgotLSAK5paw55qE55CG6K665LuO5bCR5pWw5Lq655qE5Li75byg5Yiw5LiA57uf5aSp
5LiL77yM5bm25LiN5piv5Zug5Li66L+Z5Liq55CG6K666K+05pyN5LqG5Yir5Lq65oqb5byD5pen
6KeC54K577yM6ICM5piv5Zug5Li65LiA5Luj5Lq655qE6YCd5Y6744CCCg==
--14dae9d67df4436caa04c39ae8b8
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: base64

YWJjYWJjPGJyPjxicj4tLSA8YnI+5paw55qE55CG6K665LuO5bCR5pWw5Lq655qE5Li75byg5Yiw
5LiA57uf5aSp5LiL77yM5bm25LiN5piv5Zug5Li66L+Z5Liq55CG6K666K+05pyN5LqG5Yir5Lq6
5oqb5byD5pen6KeC54K577yM6ICM5piv5Zug5Li65LiA5Luj5Lq655qE6YCd5Y6744CCPGJyPgo=
--14dae9d67df4436caa04c39ae8b8--`
		reader := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(hs)))
		hr, _ := reader.ReadMIMEHeader()
		msg := &mail.Message{
			Header: mail.Header(hr),
			Body:   bytes.NewBufferString(bd),
		}

		expect := "abcabc\n\n-- \n新的理论从少数人的主张到一统天下，并不是因为这个理论说服了别人抛弃旧观点，而是因为一代人的逝去。\n"
		content, mediatype, charset, _ := GetBody(msg, "text/plain")
		if mediatype != "text/plain" {
			t.Errorf("mediatype expect: text/plain, got: %s", mediatype)
		}
		if charset != "UTF-8" {
			t.Errorf("charset expect: UTF-8, got: %s", charset)
		}
		if content != expect {
			t.Errorf("content expect:\n%s\ngot:\n%s", expect, content)
		}
	}

	{
		hs := `MIME-Version: 1.0
Received: by 10.76.101.172 with HTTP; Tue, 26 Jun 2012 23:11:28 -0700 (PDT)
Date: Wed, 27 Jun 2012 14:11:28 +0800
Delivered-To: googollee@gmail.com
Message-ID: <CAOf82vP-CNcxcNKvRSHc_rGrNrEoTq7DLOEckuD1g-MN7LqtVg@mail.gmail.com>
Subject: test
From: Googol Lee <googollee@gmail.com>
To: =?UTF-8?B?R29vZ29sIExlZSAtIEdvb2dsZee6r+eIt+S7rO+8gemTgeihgOecn+axieWtkO+8ge+8gQ==?= <googollee@gmail.com>
Content-Type: multipart/alternative; boundary=14dae9d67df4436caa04c39ae8b8`
		bd := `--14dae9d67df4436caa04c39ae8b8
Content-Type: text/json; charset=UTF-8
Content-Transfer-Encoding: base64

YWJjYWJjCgotLSAK5paw55qE55CG6K665LuO5bCR5pWw5Lq655qE5Li75byg5Yiw5LiA57uf5aSp
5LiL77yM5bm25LiN5piv5Zug5Li66L+Z5Liq55CG6K666K+05pyN5LqG5Yir5Lq65oqb5byD5pen
6KeC54K577yM6ICM5piv5Zug5Li65LiA5Luj5Lq655qE6YCd5Y6744CCCg==
--14dae9d67df4436caa04c39ae8b8
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: base64

YWJjYWJjPGJyPjxicj4tLSA8YnI+5paw55qE55CG6K665LuO5bCR5pWw5Lq655qE5Li75byg5Yiw
5LiA57uf5aSp5LiL77yM5bm25LiN5piv5Zug5Li66L+Z5Liq55CG6K666K+05pyN5LqG5Yir5Lq6
5oqb5byD5pen6KeC54K577yM6ICM5piv5Zug5Li65LiA5Luj5Lq655qE6YCd5Y6744CCPGJyPgo=
--14dae9d67df4436caa04c39ae8b8--`
		reader := textproto.NewReader(bufio.NewReader(bytes.NewBufferString(hs)))
		hr, _ := reader.ReadMIMEHeader()
		msg := &mail.Message{
			Header: mail.Header(hr),
			Body:   bytes.NewBufferString(bd),
		}

		expect := "abcabc\n\n-- \n新的理论从少数人的主张到一统天下，并不是因为这个理论说服了别人抛弃旧观点，而是因为一代人的逝去。\n"
		content, mediatype, charset, _ := GetBody(msg, "text/plain")
		if mediatype != "text/json" {
			t.Errorf("mediatype expect: text/plain, got: %s", mediatype)
		}
		if charset != "UTF-8" {
			t.Errorf("charset expect: UTF-8, got: %s", charset)
		}
		if content != expect {
			t.Errorf("content expect:\n%s\ngot:\n%s", expect, content)
		}
	}
}
