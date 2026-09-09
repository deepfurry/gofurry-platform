package public

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestMalformedOAuthQueriesStayInsideSafeRedirect(t *testing.T) {
	app := fiber.New()
	Register(app, nil, nil, nil, Options{Environment: "test", PublicOrigin: testOrigin})
	for _, query := range []string{"code=%ZZ", "state=first&state=second", "code=first&code=second", "error=first&error=second"} {
		response, err := app.Test(httptest.NewRequest("GET", "/auth/oauth/google/callback?"+query, nil))
		if err != nil {
			t.Fatal("malformed query fixture failed")
		}
		_ = response.Body.Close()
		if response.StatusCode != 302 || response.Header.Get("Location") != testOrigin+"/login?oauth_error=provider_invalid" || response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("Referrer-Policy") != "no-referrer" {
			t.Fatal("generated parser error escaped OAuth boundary")
		}
	}
}
