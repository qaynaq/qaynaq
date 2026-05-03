package google_sheets

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/qaynaq/qaynaq/internal/connauth"
)

func (p *Processor) initSheetsService() (*sheets.Service, error) {
	p.serviceOnce.Do(func() {
		ctx := context.Background()

		var client *http.Client

		if p.serviceAccountJSON != "" {
			config, err := google.JWTConfigFromJSON(
				[]byte(p.serviceAccountJSON),
				sheets.SpreadsheetsScope,
			)
			if err != nil {
				p.serviceInitErr = fmt.Errorf("failed to parse service account JSON: %w", err)
				return
			}

			if p.delegateTo != "" {
				config.Subject = p.delegateTo
			}

			client = config.Client(ctx)
		} else if p.oauthConnection != "" {
			vp := connauth.Provider()
			if vp == nil {
				p.serviceInitErr = fmt.Errorf("vault provider not initialised")
				return
			}
			client = connauth.NewHTTPClient(ctx, vp, p.oauthConnection)
		}

		svc, err := sheets.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			p.serviceInitErr = fmt.Errorf("failed to create sheets service: %w", err)
			return
		}

		p.sheetsService = svc
	})
	return p.sheetsService, p.serviceInitErr
}
