package httpx

import (
	_ "bazaar/api/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var swaggerHandler = httpSwagger.Handler(
	httpSwagger.URL("/swagger/doc.json"),
)
