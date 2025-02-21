package tests

import (
	ssov1 "github.com/Azonnix/protos/gen/go/sso"
	"github.com/azonnix/sso/tests/suite"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	emptyAppId = 0
	appId      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg)

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appId,
	})
	require.NoError(t, err)

	loginTime := time.Now()

	token := respLogin.GetToken()
	assert.NotEmpty(t, token)

	tokenParset, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParset.Claims.(jwt.MapClaims)
	assert.True(t, ok)

	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, appId, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_DoubleRegister(t *testing.T) {
	ctx, st := suite.New(t)
	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respReg2, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg2.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	testCases := []struct {
		name        string
		email       string
		password    string
		expectError string
	}{
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			expectError: "password is required",
		},
		{
			name:        "Empty email",
			email:       "",
			password:    randomFakePassword(),
			expectError: "email is required",
		},
		{
			name:        "Empty email and password",
			email:       "",
			password:    "",
			expectError: "email is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				Email:    tc.email,
				Password: tc.password,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.expectError)
		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	testCases := []struct {
		name        string
		email       string
		password    string
		appId       int32
		expectError string
	}{
		{
			name:        "Empty password",
			email:       gofakeit.Email(),
			password:    "",
			appId:       appId,
			expectError: "password is required",
		},
		{
			name:        "Empty email",
			email:       "",
			password:    randomFakePassword(),
			appId:       appId,
			expectError: "email is required",
		},
		{
			name:        "Empty email and password",
			email:       "",
			password:    "",
			appId:       appId,
			expectError: "email is required",
		},
		{
			name:        "Empty appId",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appId:       emptyAppId,
			expectError: "appId is required",
		},
		{
			name:        "Login with invalid credentials",
			email:       gofakeit.Email(),
			password:    randomFakePassword(),
			appId:       appId,
			expectError: "invalid credentials",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				Email:    tc.email,
				Password: tc.password,
				AppId:    tc.appId,
			})
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.expectError)
		})
	}
}
