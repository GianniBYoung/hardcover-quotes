package main

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/machinebox/graphql"
)

var (
	client             *graphql.Client
	ctx                context.Context
	user_info_response *Response
	authSet            bool
)

func TestMain(m *testing.M) {

	client = graphql.NewClient(apiURL)
	ctx = context.Background()
	rawToken := os.Getenv("HARDCOVER_API_TOKEN")
	if rawToken == "" {
		log.Info("api token not set")
		authSet = false
	} else {
		authToken = strings.TrimSpace(rawToken)
		if !strings.HasPrefix(authToken, "Bearer ") {
			authToken = "Bearer " + authToken
		}
		var err error
		authSet = true
		user_info_response, err = queryUserInfo(ctx, *client)
		if err != nil {
			log.Fatal("failed to parse api", err)
		}

	}

	returnCode := m.Run()
	os.Exit(returnCode)
}

func TestVerifyUser(t *testing.T) {
	if authSet == false {
		t.Log("SKIPPING TEST - Auth token not set")
		t.Skip()
	}
	if len(user_info_response.Me) == 0 {
		log.Debug("", user_info_response.Me)
		t.Logf("User: %s", user_info_response.Me[0].Username)
		t.Fatal("unable to find username")
	}

	t.Logf("User: %s", user_info_response.Me[0].Username)

}
