// i2p-forwarding pluggable transport server.
//
// Usage (in torrc):
// 	BridgeRelay 1
// 	ORPort 9001
// 	ExtORPort 7689
// 	ServerTransportPlugin i2p exec i2p-server
//

// This transport doesn't actually transform traffic itself, instead
// it accepts traffic from I2P and forwards it back to Tor, acting
// as an "invisible bridge" to an I2P-only unlisted Tor relay. This
// makes it possible to easily set up a Tor bridge which is resistant
// to enumeration by IP address. It will appear to a network observer
// as a random I2P node. I2P uses obfuscated transports using a pair
// of protocols, one of which is based on NOISE and one of which is
// based on obfs4. So by using I2P, this pluggable transport provides
// 2 properties:

// 1. It is resistant to enumeration by IP address.
// 2. It is obfuscated and encrypted in transit by I2P.

// Since the I2P server is a pluggable transport presumably for a Tor
// relay, it's agnostic of the content it's pushing to Tor. The client
// decides how many hops it needs to take to the server. Since the
// operator presumably wants their bridge to perform well, the(hardcoded
// for now) default is to use a server-side tunnel pool composed of
// tunnels with only 1-2 hops instead of 3-4 which would provide maximum
// anonymity. This is because our goal is to resist enumeration with
// maximal performance. An option to change this value was not included
// in the release version in order to discourage abuse, but it's a one
// line change should it become necessary.

// As you can see, since I2P can implement a network interface, this
// is a very simple server. Most of the modifications that were required
// to create this have since been merged into github.com/eyedeekay/sam3,
// which is a library which allows developers to use I2P as a network
// transport for all their application development needs.
package main

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eyedeekay/sam3"
	"github.com/i2p-pt/i2p-pt/common"

	pt "git.torproject.org/pluggable-transports/goptlib.git"
)

var ptInfo pt.ServerInfo

func i2plistener(name, samaddr, keyspath string) (*sam3.StreamListener, error) {
	log.Printf("Starting and registering I2P service, please wait a couple of minutes...")
	listener, err := common.I2PSession(name, sam3.SAMDefaultAddr(samaddr), keyspath)

	if keyspath != "" {
		err = ioutil.WriteFile(keyspath+".i2p.public.txt", []byte(listener.Keys().Addr().Base32()), 0644)
		if err != nil {
			log.Fatalf("error storing I2P base32 address in adjacent text file, %s", err)
		}
	}
	log.Printf("Listening on: %s", listener.Addr().Base32())
	return listener.Listen()
}

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

func handler(conn net.Conn) error {
	defer conn.Close()

	or, err := pt.DialOr(&ptInfo, conn.RemoteAddr().String(), "i2p")
	if err != nil {
		return err
	}
	defer or.Close()

	copyLoop(conn, or)

	return nil
}

func acceptLoop(ln net.Listener) error {
	defer ln.Close()
	for {
		conn, err := ln.Accept()
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

	ptInfo, err = pt.ServerSetup(nil)
	if err != nil {
		os.Exit(1)
	}

	listeners := make([]net.Listener, 0)
	for _, bindaddr := range ptInfo.Bindaddrs {
		switch bindaddr.MethodName {
		case "i2p":
			ln, err := i2plistener("i2p-tor-bridge", "127.0.0.1:7656", "i2p-tor-bridge")
			if err != nil {
				pt.SmethodError(bindaddr.MethodName, err.Error())
				break
			}
			go acceptLoop(ln)
			pt.Smethod(bindaddr.MethodName, ln.Addr())
			listeners = append(listeners, ln)
		default:
			pt.SmethodError(bindaddr.MethodName, "no such method")
		}
	}
	pt.SmethodsDone()

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
