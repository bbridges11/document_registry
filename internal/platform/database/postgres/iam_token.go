package postgres

import (
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// iamTokenGenerator generates and caches RDS IAM auth tokens
type iamTokenGenerator struct {
	endpoint string // host:port
	region   string
	dbUser   string

	mu        sync.RWMutex
	token     string
	expiresAt time.Time
}

// Token cache duration - refresh tokens after 10 minutes (before 15min expiry)
const tokenCacheDuration = 10 * time.Minute

// newIAMTokenGenerator creates a new IAM token generator
func newIAMTokenGenerator(endpoint, region, dbUser string) *iamTokenGenerator {
	return &iamTokenGenerator{
		endpoint: endpoint,
		region:   region,
		dbUser:   dbUser,
	}
}

// getToken returns a valid IAM auth token, generating a new one if needed
func (g *iamTokenGenerator) getToken(ctx context.Context) (token string, err error) {
	defer err2.Handle(&err)

	// Check if cached token is still valid
	g.mu.RLock()
	if time.Now().Before(g.expiresAt) && g.token != "" {
		token := g.token
		g.mu.RUnlock()
		return token, nil
	}
	g.mu.RUnlock()

	// Need to generate new token
	g.mu.Lock()
	defer g.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine may have refreshed)
	if time.Now().Before(g.expiresAt) && g.token != "" {
		return g.token, nil
	}

	// Load AWS config - this automatically uses:
	// 1. ECS Task IAM Role (when running in ECS)
	// 2. EC2 Instance Profile (when running on EC2)
	// 3. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
	// 4. Shared credentials file (~/.aws/credentials)
	cfg := try.To1(config.LoadDefaultConfig(ctx, config.WithRegion(g.region)))

	// Generate IAM auth token
	// This creates a temporary password that's valid for 15 minutes
	authToken := try.To1(auth.BuildAuthToken(
		ctx,
		g.endpoint, // e.g., "my-cluster.cluster-abc.us-east-1.rds.amazonaws.com:6160"
		g.region,   // e.g., "us-east-1"
		g.dbUser,   // e.g., "iam_db_user"
		cfg.Credentials,
	))

	// Cache the token
	g.token = authToken
	g.expiresAt = time.Now().Add(tokenCacheDuration)

	return authToken, nil
}
