package main

import (
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/signal"

	"github.com/k0kubun/pp"
	"github.com/phalaaxx/milter"
)

type testMilter struct {
	authAuthen string
}

var count map[string]int

/* Header parses message headers one by one */
func (t *testMilter) Header(name, value string, m *milter.Modifier) (milter.Response, error) {
	/*fmt.Println("--------------------------------------------------------------------")
	fmt.Println("Header")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("name:%s\n", name)
	pp.Println(m)*/
	return milter.RespContinue, nil
}

/* MailFrom is called on envelope from address */
func (t *testMilter) MailFrom(from string, m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("MailFrom")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("from:%s\n", from)
	pp.Println(m)
	if val, ok := m.Macros["{auth_authen}"]; ok {
		t.authAuthen = val
		pp.Println(val)
	}
	return milter.RespContinue, nil
}

func (t *testMilter) Connect(host string, family string, port uint16, addr net.IP, m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("Connect")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("host:%s family:%s port:%d addr:%s\n", host, family, port, addr.String())
	pp.Println(m)
	pp.Println(t)
	return milter.RespContinue, nil
}

func (t *testMilter) Helo(name string, m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("Helo")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("name:%s\n", name)
	pp.Println(m)
	return milter.RespContinue, nil
}

func (t *testMilter) RcptTo(rcptTo string, m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("RcptTo")
	fmt.Println("--------------------------------------------------------------------")
	fmt.Printf("rcpto:%s\n", rcptTo)
	pp.Println(m)
	pp.Println(t)
	if t.authAuthen != "" {
		if _, ok := count[t.authAuthen]; ok {
			count[t.authAuthen]++
		} else {
			count[t.authAuthen] = 1
		}
	}

	pp.Println(count)
	return milter.RespContinue, nil
}
func (t *testMilter) Headers(headers textproto.MIMEHeader, m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("Headers")
	fmt.Println("--------------------------------------------------------------------")
	pp.Println(headers)
	pp.Println(m)
	return milter.RespContinue, nil
}

func (t *testMilter) BodyChunk(chunk []byte, m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("BodyChunk")
	fmt.Println("--------------------------------------------------------------------")
	pp.Println(string(chunk))
	pp.Println(m)
	return milter.RespContinue, nil
}

func (t *testMilter) Body(m *milter.Modifier) (milter.Response, error) {
	fmt.Println("--------------------------------------------------------------------")
	fmt.Println("Body")
	fmt.Println("--------------------------------------------------------------------")
	pp.Println(m)
	return milter.RespContinue, nil
}

func runServer(socket net.Listener) {
	if err := milter.RunServer(socket, func() (milter.Milter, milter.OptAction, milter.OptProtocol) {
		return &testMilter{},
			milter.OptAddHeader | milter.OptChangeHeader,
			milter.OptNoBody
	}); err != nil {
		log.Fatal(err)
	}
}

func main() {
	count = make(map[string]int)
	file := "/var/spool/postfix/testmilter.sock"
	sock, err := net.Listen("unix", file)
	if err != nil {
		log.Fatal(err)
	}
	defer sock.Close()

	// set mode 0660 for unix domain sockets
	if err := os.Chmod(file, 0666); err != nil {
		log.Fatal(err)
	}
	// remove socket on exit
	defer os.Remove(file)

	go runServer(sock)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	switch <-sig {
	default:
		os.Remove(file)
	}
}
