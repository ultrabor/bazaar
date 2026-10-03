package httpx

import (
	_ "bazaar/api/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var swaggerHandler = httpSwagger.Handler(
	httpSwagger.URL("/swagger/doc.json"),
	httpSwagger.UIConfig(map[string]string{
		"requestInterceptor": `(req) => {
			const token = req.headers?.Authorization;
			if (token && !token.startsWith("Bearer ")) {
				req.headers.Authorization = "Bearer " + token;
			}
			return req;
		}`,
	}),
)
