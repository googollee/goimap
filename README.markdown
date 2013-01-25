goimap
======

IMAP Client for go

ATTENTION
---------

- Not fully implemented.
- Only tested with GMail.

Usage
-----

    package main

    import (
        "fmt"
        "io/ioutil"
        "imap"
        "net"
    )

    func get1st(a, b interface{}) interface{} {
        return a
    }

    func main() {
        conn, _ := net.Dial("tcp", "imap.gmail.com:993")
        client, _ := imap.NewClient(conn, "imap.gmail.com")
        defer client.Close()

        _ = client.Login("mail@gmail.com", "password")
        client.Select(imap.Inbox)
        ids, _ := client.Search("unseen")
        fmt.Println(ids)

        for _, id := range ids {
            client.StoreFlag(id, imap.Seen)

            msg, _ := client.GetMessage(id)

            fmt.Println("To:", get1st(msg.Header.AddressList("To")))
            fmt.Println("From:", get1st(msg.Header.AddressList("From")))
            from, _ := msg.Header.AddressList("To")
            fmt.Println("From:", from[0].Name)
            fmt.Println("Subject:", msg.Header["Subject"])
            fmt.Println("Date:", get1st(msg.Header.Date()))
            body, _ := ioutil.ReadAll(msg.Body)
            fmt.Println("body:\n", string(body))
        }
        client.Logout()
    }
