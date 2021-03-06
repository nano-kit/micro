package client

import (
	"context"

	ccli "github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/auth"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/micro/v2/client/cli/util"
	cliutil "github.com/micro/micro/v2/client/cli/util"
	"github.com/micro/micro/v2/internal/config"
)

// New returns a wrapped grpc client which will inject the
// token and namespace found in config into each request
func New(ctx *ccli.Context, opts ...Option) client.Client {
	env := cliutil.GetEnv(ctx)
	token, _ := config.Get("micro", "auth", env.Name, "token")
	ns, _ := config.Get("micro", "auth", env.Name, "namespace")
	a := &wrapper{grpc.NewClient(), ns, token, env.Name, ctx}
	for _, o := range opts {
		o(a)
	}
	return a
}

type wrapper struct {
	client.Client
	namespace string
	token     string
	env       string
	ctx       *ccli.Context
}

func (a *wrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	if len(a.token) > 0 {
		ctx = metadata.Set(ctx, "Authorization", auth.BearerScheme+a.token)
	}
	if len(a.env) > 0 && !util.IsLocal(a.ctx) && !util.IsServer(a.ctx) {
		// @todo this is temporarily removed because multi tenancy is not there yet
		// and the moment core and non core services run in different environments, we
		// get issues. To test after `micro env add mine 127.0.0.1:8081` do,
		// `micro run github.com/crufter/micro-services/logspammer` works but
		// `micro -env=mine run github.com/crufter/micro-services/logspammer` is broken.
		// Related ticket https://github.com/micro/development/issues/193
		//
		// env := strings.ReplaceAll(a.env, "/", "-")
		// ctx = metadata.Set(ctx, "Micro-Namespace", env)
	}
	if len(a.namespace) > 0 {
		ctx = metadata.Set(ctx, "Micro-Namespace", a.namespace)
	}
	return a.Client.Call(ctx, req, rsp, opts...)
}

type Option func(*wrapper)

func WithNamespace(ns string) Option {
	return func(a *wrapper) {
		a.namespace = ns
	}
}
