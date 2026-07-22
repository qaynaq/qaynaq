package bloblang

import (
	"fmt"

	"github.com/warpstreamlabs/bento/public/bloblang"

	"github.com/qaynaq/qaynaq/internal/connauth"
)

// qaynaq_connection_token is not meant to be written by hand: the coordinator's
// config builder injects it into HTTP component headers when a user picks a
// connection. Registered globally so both coordinator (lint, try) and worker
// (execution) can parse it; execution requires a vault provider.
func init() {
	spec := bloblang.NewPluginSpec().
		Category("General").
		Description("Returns a fresh access token for a Qaynaq connection.").
		Param(bloblang.NewStringParam("name").Description("The connection name."))

	err := bloblang.RegisterFunctionV2("qaynaq_connection_token", spec, func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		name, err := args.GetString("name")
		if err != nil {
			return nil, err
		}
		return func() (any, error) {
			vp := connauth.Provider()
			if vp == nil {
				return nil, fmt.Errorf("connection %q tokens are not available in this context", name)
			}
			tok, err := vp.GetAccessToken(name)
			if err != nil {
				return nil, fmt.Errorf("failed to get access token for connection %q: %w", name, err)
			}
			return tok.AccessToken, nil
		}, nil
	})
	if err != nil {
		panic(err)
	}
}
