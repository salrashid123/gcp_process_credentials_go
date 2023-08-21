// Copyright 2020 Google LLC.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

type ExternalTokenConfig struct {
	Env     []string
	Command string
	Args    []string
	Parser  func([]byte) (ExternalTokenResponse, error)
}

// https://github.com/golang/oauth2/blob/master/internal/token.go#L33
type ExternalTokenResponse struct {
	Token     string `json:"access_token"`
	TokenType string `json:"token_type,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
}

const ()

func ExternalTokenSource(tokenConfig *ExternalTokenConfig) (oauth2.TokenSource, error) {

	if tokenConfig.Command == "" {
		return nil, fmt.Errorf("oauth2/google: Command cannot be nil")
	}

	return &externalTokenSource{
		refreshMutex:  &sync.Mutex{},
		externalToken: nil,
		env:           tokenConfig.Env,
		command:       tokenConfig.Command,
		args:          tokenConfig.Args,
		parser:        tokenConfig.Parser,
	}, nil
}

type externalTokenSource struct {
	refreshMutex  *sync.Mutex
	externalToken *oauth2.Token
	env           []string
	command       string
	args          []string
	parser        func([]byte) (ExternalTokenResponse, error)
}

func (ts *externalTokenSource) Token() (*oauth2.Token, error) {

	ts.refreshMutex.Lock()
	defer ts.refreshMutex.Unlock()

	if ts.externalToken.Valid() {
		return ts.externalToken, nil
	}
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(2)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, ts.command, ts.args...)
	cmd.Env = append(os.Environ(), ts.env...)

	stdout := &bytes.Buffer{}
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Unable to run command %v", err)
	}

	if ts.parser != nil {
		s, err := ts.parser(stdout.Bytes())
		if err != nil {
			return nil, fmt.Errorf("Unable to run Custom Parser %v", err)
		}
		return &oauth2.Token{
			AccessToken: s.Token,
			Expiry:      time.Now().Add(time.Duration(s.ExpiresIn)), // time.Now().UTC().Add(time.Duration(s.ExpiresIn)),
			TokenType:   s.TokenType,
		}, nil
	} else {
		resp := &ExternalTokenResponse{}
		err = json.Unmarshal(stdout.Bytes(), resp)
		if err != nil {
			return nil, err
		} else {
			return &oauth2.Token{
				AccessToken: resp.Token,
				TokenType:   resp.TokenType,
				Expiry:      time.Now().Add(time.Duration(resp.ExpiresIn)),
			}, nil
		}
	}
}
