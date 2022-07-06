// i2p-forwarding pluggable transport client.
//
// Usage (in torrc):
// 	UseBridges 1
// 	Bridge i2p i2pbase32addressesarefiftytwocharacterslongenoughok.b32.i2p
// 	ClientTransportPlugin i2p exec i2p-client

// Like the server transport, this pluggable transport outsources the task
// of transforming Tor traffic to I2P instead of performing the transformation
// itself. As such, it requires there to be an I2P router on the machine which
// uses it, although that router may operate in Hidden Mode further enhancing
// it's resistance to enumeration by not routing traffic for others and not
// participating in the global routing table of I2P routers(The NetDB). It is
// recommended that people in sensitive or restricted locations use Java I2P
// so that hidden mode is enabled automatically based on their location.

// Also like the server transport, the primary goals of this transport are
// enumeration resistance and obfuscation while your Tor traffic reaches a
// bridge using I2P. As such, it is optimized for performance while accomplishing
// these goals. It uses 1-2 hops over an I2P tunnel pool to obfuscate traffic and
// make it impossible for the bridge to know the location of the client.

// -- STOP for page generation -- //

// see also /i2p-server/i2p-server.go

package main

import (
	"io"
	"io/ioutil"
	"net"

	//"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pt "git.torproject.org/pluggable-transports/goptlib.git"
	"github.com/eyedeekay/sam3"
	"github.com/i2p-pt/i2p-pt/common"
)

var ptInfo pt.ClientInfo

var session *sam3.StreamSession

func copyLoop(a, b net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		io.Copy(b, a)
		wg.Done()
	}()
	go func() {
		io.Copy(a, b)
		wg.Done()
	}()

	wg.Wait()
}

func handler(conn *pt.SocksConn) error {
	defer conn.Close()
	//remote, err := net.Dial("tcp", conn.Req.Target)
	remote, err := session.Dial("tcp", conn.Req.Target)
	if err != nil {
		conn.Reject()
		return err
	}
	defer remote.Close()
	err = conn.Grant(remote.RemoteAddr().(*net.TCPAddr))
	if err != nil {
		return err
	}

	copyLoop(conn, remote)

	return nil
}

func acceptLoop(ln *pt.SocksListener) error {
	defer ln.Close()
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				continue
			}
			return err
		}
		go handler(conn)
	}
}

func main() {
	var err error
	session, err = common.I2PSession("i2p-tor-client", sam3.SAMDefaultAddr(""), "i2p-tor-client")
	if err != nil {
		os.Exit(1)
	}

	ptInfo, err = pt.ClientSetup(nil)
	if err != nil {
		os.Exit(1)
	}

	if ptInfo.ProxyURL != nil {
		pt.ProxyError("proxy is not supported")
		os.Exit(1)
	}

	listeners := make([]net.Listener, 0)
	for _, methodName := range ptInfo.MethodNames {
		switch methodName {
		case "i2p":
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				pt.CmethodError(methodName, err.Error())
				break
			}
			go acceptLoop(ln)
			pt.Cmethod(methodName, ln.Version(), ln.Addr())
			listeners = append(listeners, ln)
		default:
			pt.CmethodError(methodName, "no such method")
		}
	}
	pt.CmethodsDone()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	if os.Getenv("TOR_PT_EXIT_ON_STDIN_CLOSE") == "1" {
		// This environment variable means we should treat EOF on stdin
		// just like SIGTERM: https://bugs.torproject.org/15435.
		go func() {
			io.Copy(ioutil.Discard, os.Stdin)
			sigChan <- syscall.SIGTERM
		}()
	}

	// wait for a signal
	<-sigChan

	// signal received, shut down
	for _, ln := range listeners {
		ln.Close()
	}
}
