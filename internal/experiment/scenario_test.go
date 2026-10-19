package experiment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var devTiming = SystemTiming{ReservationTTL: 10 * time.Second, CleanerInterval: 5 * time.Second, OutboxInterval: 5 * time.Second}

func TestBuildScenario_PaymentOutage(t *testing.T) {
	s, err := BuildScenario(ScenarioPaymentOutage, devTiming, 0)
	require.NoError(t, err)

	assert.Equal(t, "payment-service", s.Service)
	assert.Equal(t, "0", s.FaultEnv["FAKE_BANK_SUCCESS_CHANCE"])
	assert.True(t, s.Expects(MetricConversionRate))
	assert.True(t, s.Expects(MetricPaymentSuccessRate))
	assert.False(t, s.Expects(MetricAbandonmentRateExpired))
}

func TestBuildScenario_BankTimeoutDefaultsDelayToTwiceTTL(t *testing.T) {
	s, err := BuildScenario(ScenarioBankTimeout, devTiming, 0)
	require.NoError(t, err)

	assert.Equal(t, "20s", s.FaultEnv["FAKE_BANK_MIN_DELAY"])
	assert.Equal(t, "20s", s.FaultEnv["FAKE_BANK_MAX_DELAY"])
	assert.True(t, s.Expects(MetricAbandonmentRateExpired))
	// Expired outcomes cannot appear earlier than TTL + cleaner tick + outbox tick.
	assert.Greater(t, s.Timeout, 20*time.Second)
}

func TestBuildScenario_BankTimeoutRejectsDelayWithinTTL(t *testing.T) {
	_, err := BuildScenario(ScenarioBankTimeout, devTiming, 5*time.Second)
	assert.Error(t, err)
}

func TestBuildScenario_Unknown(t *testing.T) {
	_, err := BuildScenario("nope", devTiming, 0)
	assert.Error(t, err)
}

func TestLoadEnvFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(path, []byte(`
# comment
RESERVATION_TTL=15m
USER_DB_USER=postgres
USER_DB_PASSWORD="password"
USER_DB_PORT_EXTERNAL=5433
USER_DB_NAME=user_db
EMPTY=
`), 0o644))

	env, err := LoadEnvFile(path)
	require.NoError(t, err)

	ttl, err := env.Duration("RESERVATION_TTL")
	require.NoError(t, err)
	assert.Equal(t, 15*time.Minute, ttl)

	url, err := env.PostgresURL("USER_DB")
	require.NoError(t, err)
	assert.Equal(t, "postgres://postgres:password@localhost:5433/user_db?sslmode=disable", url)

	_, err = env.Get("EMPTY")
	assert.Error(t, err)
}
