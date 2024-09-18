package config

import "os"

var JwtKey = []byte(os.Getenv("JWT_SECRET_KEY")) // Replace with a strong secret key
var JwtRefreshKey = []byte(os.Getenv("JWT_REFRESH_KEY")) // Secret for refresh token