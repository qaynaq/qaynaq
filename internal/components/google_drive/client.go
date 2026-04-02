package google_drive

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/qaynaq/qaynaq/internal/connection"
)

func (p *Processor) initDriveService() (*drive.Service, error) {
	p.serviceOnce.Do(func() {
		ctx := context.Background()

		var client *http.Client

		if p.serviceAccountJSON != "" {
			config, err := google.JWTConfigFromJSON(
				[]byte(p.serviceAccountJSON),
				drive.DriveScope,
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
			var connData connection.ConnectionData
			if err := json.Unmarshal([]byte(p.oauthConnection), &connData); err != nil {
				p.serviceInitErr = fmt.Errorf("failed to parse OAuth connection data: %w", err)
				return
			}

			endpoint, err := connection.GetEndpoint(connData.Provider)
			if err != nil {
				p.serviceInitErr = fmt.Errorf("failed to get OAuth endpoint: %w", err)
				return
			}

			oauth2Config := &oauth2.Config{
				ClientID:     connData.ClientID,
				ClientSecret: connData.ClientSecret,
				Endpoint:     endpoint,
			}

			token := &oauth2.Token{
				AccessToken:  connData.Token.AccessToken,
				RefreshToken: connData.Token.RefreshToken,
				TokenType:    connData.Token.TokenType,
				Expiry:       connData.Token.Expiry,
			}

			client = oauth2.NewClient(ctx, oauth2Config.TokenSource(ctx, token))
		}

		svc, err := drive.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			p.serviceInitErr = fmt.Errorf("failed to create drive service: %w", err)
			return
		}

		p.driveService = svc
	})
	return p.driveService, p.serviceInitErr
}
