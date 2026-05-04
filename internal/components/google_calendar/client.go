package google_calendar

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/qaynaq/qaynaq/internal/connauth"
)

func (p *Processor) initCalendarService() (*calendar.Service, error) {
	p.serviceOnce.Do(func() {
		ctx := context.Background()

		var client *http.Client

		if p.serviceAccountJSON != "" {
			config, err := google.JWTConfigFromJSON(
				[]byte(p.serviceAccountJSON),
				calendar.CalendarScope,
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

		svc, err := calendar.NewService(ctx, option.WithHTTPClient(client))
		if err != nil {
			p.serviceInitErr = fmt.Errorf("failed to create calendar service: %w", err)
			return
		}

		p.calendarService = svc
	})
	return p.calendarService, p.serviceInitErr
}
