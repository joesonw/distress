package main

import (
	"io"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	http.ListenAndServe(":9100", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// handle error
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err == io.EOF {
					return
				}
				if err != nil {
					panic(err)
				}
				err = wsutil.WriteServerMessage(conn, op, append([]byte("you said "), msg...))
				if err != nil {
					panic(err)
				}
			}
		}()
	}))
}
