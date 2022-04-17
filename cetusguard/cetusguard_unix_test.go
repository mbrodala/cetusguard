//go:build aix || android || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || solaris

package cetusguard

import (
	"context"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCetusGuardSocketAllowedReq(t *testing.T) {
	tc := &testCase{
		daemonListenerFunc: socketDaemonListener,
		daemonFunc:         socketDaemon,
		backendFunc:        socketBackend,
		frontendFunc:       socketFrontend,
		clientFunc:         socketClient,
	}

	defer tc.setup(t)()
	tc.daemon.Handler = http.HandlerFunc(httpDaemonHandler)
	tc.server.Rules = testRules

	ready := make(chan any, 1)
	go func() {
		err := tc.server.Start(ready)
		if err != nil {
			t.Error(err)
		}
	}()
	<-ready

	addr, err := tc.server.Addr()
	if err != nil {
		t.Fatal(err)
	}

	req, err := httpClientReq("http", addr.String())
	if err != nil {
		t.Fatal(err)
	}

	res, err := tc.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("res.StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}

	msg, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(msg) != "PONG" {
		t.Fatalf(`msg = "%s", want "%s"`, msg, "PONG")
	}

	err = tc.server.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCetusGuardSocketAllowedStreamReq(t *testing.T) {
	tc := &testCase{
		daemonListenerFunc: socketDaemonListener,
		daemonFunc:         socketDaemon,
		backendFunc:        socketBackend,
		frontendFunc:       socketFrontend,
		clientFunc:         socketClient,
	}

	defer tc.setup(t)()
	tc.daemon.Handler = http.HandlerFunc(httpDaemonHandler)
	tc.server.Rules = testRules

	ready := make(chan any, 1)
	go func() {
		err := tc.server.Start(ready)
		if err != nil {
			t.Error(err)
		}
	}()
	<-ready

	addr, err := tc.server.Addr()
	if err != nil {
		t.Fatal(err)
	}

	req, err := httpClientReq("http", addr.String())
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Upgrade", "tcp")
	req.Header.Set("Connection", "Upgrade")

	res, err := tc.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("res.StatusCode = %d, want %d", res.StatusCode, http.StatusSwitchingProtocols)
	}

	msg := make([]byte, 4)
	_, err = io.ReadFull(res.Body, msg)
	if err != nil {
		t.Fatal(err)
	}

	if string(msg) != "PONG" {
		t.Fatalf(`msg = "%s", want "%s"`, msg, "PONG")
	}

	err = tc.server.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCetusGuardSocketDeniedReq(t *testing.T) {
	tc := &testCase{
		daemonListenerFunc: socketDaemonListener,
		daemonFunc:         socketDaemon,
		backendFunc:        socketBackend,
		frontendFunc:       socketFrontend,
		clientFunc:         socketClient,
	}

	defer tc.setup(t)()
	tc.daemon.Handler = http.HandlerFunc(httpDaemonHandler)

	ready := make(chan any, 1)
	go func() {
		err := tc.server.Start(ready)
		if err != nil {
			t.Error(err)
		}
	}()
	<-ready

	addr, err := tc.server.Addr()
	if err != nil {
		t.Fatal(err)
	}

	req, err := httpClientReq("http", addr.String())
	if err != nil {
		t.Fatal(err)
	}

	res, err := tc.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusForbidden {
		t.Fatalf("res.StatusCode = %d, want %d", res.StatusCode, http.StatusForbidden)
	}

	err = tc.server.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCetusGuardSocketTlsAuthBackendReq(t *testing.T) {
	tc := &testCase{
		daemonListenerFunc: tcpDaemonListener,
		daemonFunc:         tlsAuthDaemon,
		backendFunc:        tlsAuthBackend,
		frontendFunc:       socketFrontend,
		clientFunc:         socketClient,
	}

	defer tc.setup(t)()
	tc.daemon.Handler = http.HandlerFunc(httpDaemonHandler)
	tc.server.Rules = testRules

	ready := make(chan any, 1)
	go func() {
		err := tc.server.Start(ready)
		if err != nil {
			t.Error(err)
		}
	}()
	<-ready

	addr, err := tc.server.Addr()
	if err != nil {
		t.Fatal(err)
	}

	req, err := httpClientReq("http", addr.String())
	if err != nil {
		t.Fatal(err)
	}

	res, err := tc.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("res.StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}

	msg, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(msg) != "PONG" {
		t.Fatalf(`msg = "%s", want "%s"`, msg, "PONG")
	}

	err = tc.server.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func TestCetusGuardTlsAuthSocketBackendReq(t *testing.T) {
	tc := &testCase{
		daemonListenerFunc: socketDaemonListener,
		daemonFunc:         socketDaemon,
		backendFunc:        socketBackend,
		frontendFunc:       tlsAuthFrontend,
		clientFunc:         tlsAuthClient,
	}

	defer tc.setup(t)()
	tc.daemon.Handler = http.HandlerFunc(httpDaemonHandler)
	tc.server.Rules = testRules

	ready := make(chan any, 1)
	go func() {
		err := tc.server.Start(ready)
		if err != nil {
			t.Error(err)
		}
	}()
	<-ready

	addr, err := tc.server.Addr()
	if err != nil {
		t.Fatal(err)
	}

	req, err := httpClientReq("https", addr.String())
	if err != nil {
		t.Fatal(err)
	}

	res, err := tc.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("res.StatusCode = %d, want %d", res.StatusCode, http.StatusOK)
	}

	msg, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(msg) != "PONG" {
		t.Fatalf(`msg = "%s", want "%s"`, msg, "PONG")
	}

	err = tc.server.Stop()
	if err != nil {
		t.Fatal(err)
	}
}

func socketDaemonListener(tmpdir string) (net.Listener, error) {
	listener, err := net.Listen("unix", filepath.Join(tmpdir, "d"))
	if err != nil {
		return nil, err
	}

	return listener, nil
}

func socketDaemon() (*http.Server, error) {
	server, err := plainDaemon()
	if err != nil {
		return nil, err
	}

	return server, nil
}

func socketBackend(listener net.Listener, tmpdir string) (*Backend, error) {
	backend, err := plainBackend(listener, tmpdir)
	if err != nil {
		return nil, err
	}

	return backend, nil
}

func socketFrontend(tmpdir string) (*Frontend, error) {
	frontend, err := plainFrontend(tmpdir)
	if err != nil {
		return nil, err
	}

	frontend.Addr = "unix://" + filepath.Join(tmpdir, "c")

	return frontend, nil
}

func socketClient() (*http.Client, error) {
	client, err := plainClient()
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 90 * time.Second,
	}
	transport := client.Transport.(*http.Transport)
	transport.DialContext = func(ctx context.Context, _ string, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "unix", addr[:strings.LastIndex(addr, ":")])
	}

	return client, nil
}
