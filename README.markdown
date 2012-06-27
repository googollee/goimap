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
    )

    func get1st(a, b interface{}) interface{} {
        return a
    }

    func main() {
        conn, _ := imap.NewClient("imap.gmail.com:993")
        defer conn.Close()

        _ = conn.Login("mail@gmail.com", "password")
        conn.Select(imap.Inbox)
        ids, _ := conn.Search("unseen")
        fmt.Println(ids)

        for _, id := range ids {
            conn.StoreFlag(id, imap.Seen)

            msg, _ := conn.GetMessage(id)

            fmt.Println("To:", get1st(msg.Header.AddressList("To")))
            fmt.Println("From:", get1st(msg.Header.AddressList("From")))
            from, _ := msg.Header.AddressList("To")
            fmt.Println("From:", from[0].Name)
            fmt.Println("Subject:", msg.Header["Subject"])
            fmt.Println("Date:", get1st(msg.Header.Date()))
            body, _ := ioutil.ReadAll(msg.Body)
            fmt.Println("body:\n", string(body))
        }
        conn.Logout()
    }
