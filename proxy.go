package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"errors"
	"context"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"time"
	"fmt"
	"golang.org/x/sync/errgroup"
)

type GProxy struct {
	conf        *GProxyConf
	serverHttp  *http.Server
	serverHttps *http.Server
	client      *http.Client
	httpRoutes  []*HttpRoute
	httpsRoutes []*HttpRoute
}

func NewGProxy(confPath string) (*GProxy, error) {
	cfgData, e := ioutil.ReadFile(confPath);
	if e != nil {
		return nil, e
	}
	conf := &GProxyConf{}
	e = json.Unmarshal(cfgData, conf)
	if e != nil {
		return nil, e
	}
	proxy := new(GProxy)
	proxy.conf = conf
	proxy.initLog()
	e = proxy.initClient()
	if e != nil {
		return nil, e
	}
	e = proxy.initServer()
	if e != nil {
		return nil, e
	}
	return proxy, nil
}

func (p *GProxy) Start() error {
	var g errgroup.Group
	if p.serverHttp != nil {
		g.Go(func() error {
			err := p.serverHttp.ListenAndServe()
			if err != nil {
				log.Println(err)
			}
			return err
		})
	}
	if p.serverHttps != nil {
		g.Go(func() error {
			err := p.serverHttps.ListenAndServeTLS(p.conf.Https.CertFile, p.conf.Https.KeyFile)
			if err != nil {
				log.Println(err)
			}
			return err
		})
	}
	return g.Wait()
}

func (p *GProxy) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return p.serverHttp.Shutdown(ctx)
}

func (p *GProxy) initLog() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   p.conf.Log,
		MaxSize:    64,
		MaxBackups: 32,
		MaxAge:     1,
		LocalTime:  true,
		Compress:   true,
	})

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

func (p *GProxy) initClient() error {
	connectionConf := p.conf.Http.Connection
	tr := &http.Transport{
		MaxIdleConns:       connectionConf.MaxIdleConns,
		IdleConnTimeout:    time.Duration(connectionConf.IdleConnTimeout) * time.Second,
		DisableCompression: false,
	}
	p.client = &http.Client{
		Transport: tr,
		Timeout:   time.Duration(connectionConf.Timeout) * time.Second,
	}
	return nil
}

func (p *GProxy) initServer() error {
	if e := p.initHttpServer(); e != nil {
		return e
	}
	if e := p.initHttpsServer(); e != nil {
		return e
	}
	return nil
}

func (p *GProxy) initHttpServer() error {
	httpConf := p.conf.Http
	serverConf := httpConf.Server
	listenAddr := fmt.Sprintf(":%d", httpConf.Server.Port)
	p.serverHttp = &http.Server{
		Addr:           listenAddr,
		ReadTimeout:    time.Duration(serverConf.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(serverConf.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	p.httpRoutes = make([]*HttpRoute, len(serverConf.Locations))
	for _, location := range serverConf.Locations {
		if location.Path == "" {
			return errors.New("Empty location path")
		}
		if location.StaticRoot != "" {
			fs := http.FileServer(http.Dir(location.StaticRoot))
			http.Handle(location.Path, fs)
		} else {
			if location.Destination == "" {
				return errors.New(fmt.Sprint("Empty destination for location [%s]", location.Path))
			}
			route := &HttpRoute{
				path:     location.Path,
				location: location,
				proxy:    p,
			}
			http.Handle(location.Path, http.HandlerFunc(route.httpHandler))
			p.httpRoutes = append(p.httpRoutes, route)
		}
	}
	return nil
}

func (p *GProxy) initHttpsServer() error {
	httpConf := p.conf.Https
	serverConf := httpConf.Server
	listenAddr := fmt.Sprintf(":%d", httpConf.Server.Port)
	p.serverHttps = &http.Server{
		Addr:           listenAddr,
		ReadTimeout:    time.Duration(serverConf.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(serverConf.WriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	p.httpsRoutes = make([]*HttpRoute, len(serverConf.Locations))
	for _, location := range serverConf.Locations {
		if location.Path == "" {
			return errors.New("Empty location path")
		}
		if location.StaticRoot != "" {
			fs := http.FileServer(http.Dir(location.StaticRoot))
			http.Handle(location.Path, fs)
		} else {
			if location.Destination == "" {
				return errors.New(fmt.Sprint("Empty destination for location [%s]", location.Path))
			}
			route := &HttpRoute{
				path:     location.Path,
				location: location,
				proxy:    p,
			}
			http.Handle(location.Path, http.HandlerFunc(route.httpHandler))
			p.httpsRoutes = append(p.httpRoutes, route)
		}
	}
	return nil
}
