package httpx

import (
	"fmt"
	"net/http"
)

const openAPISpecYAML = `openapi: 3.0.3
info:
  title: CityHawk Backend API
  version: 1.0.0
  description: HTTP API for CityHawk backend service.
servers:
  - url: http://localhost:8080
tags:
  - name: Health
  - name: Auth
  - name: User
  - name: Places
paths:
  /api/health:
    get:
      tags: [Health]
      summary: Health check
      responses:
        "200":
          description: Service is alive
          content:
            text/plain:
              schema:
                type: string
                example: ok
  /api/auth/register:
    post:
      tags: [Auth]
      summary: Register user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RegisterRequest'
      responses:
        "201":
          description: Registration successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/RegisterResponse'
        "400":
          description: Validation or JSON error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "409":
          description: Email already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/login:
    post:
      tags: [Auth]
      summary: Login user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
      responses:
        "200":
          description: Login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/LoginResponse'
        "400":
          description: Validation or JSON error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "401":
          description: Invalid credentials
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/google/login:
    get:
      tags: [Auth]
      summary: Start Google OAuth flow
      responses:
        "302":
          description: Redirect to Google OAuth consent page
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/google/callback:
    get:
      tags: [Auth]
      summary: Google OAuth callback
      parameters:
        - in: query
          name: state
          required: true
          schema:
            type: string
        - in: query
          name: code
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Google login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MessageResponse'
        "400":
          description: Missing oauth code
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "401":
          description: Invalid oauth state/code or profile fetch failure
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/yandex/login:
    get:
      tags: [Auth]
      summary: Start Yandex OAuth flow
      responses:
        "302":
          description: Redirect to Yandex OAuth consent page
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/yandex/callback:
    get:
      tags: [Auth]
      summary: Yandex OAuth callback
      parameters:
        - in: query
          name: state
          required: true
          schema:
            type: string
        - in: query
          name: code
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Yandex login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MessageResponse'
        "400":
          description: Missing oauth code
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "401":
          description: Invalid oauth state/code or profile fetch failure
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/vk/login:
    get:
      tags: [Auth]
      summary: Start VK OAuth flow
      responses:
        "302":
          description: Redirect to VK OAuth consent page
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/vk/callback:
    get:
      tags: [Auth]
      summary: VK OAuth callback
      parameters:
        - in: query
          name: state
          required: true
          schema:
            type: string
        - in: query
          name: code
          required: false
          schema:
            type: string
        - in: query
          name: error
          required: false
          schema:
            type: string
      responses:
        "200":
          description: VK login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MessageResponse'
        "400":
          description: Missing oauth code
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "401":
          description: Invalid oauth state/code or oauth denied
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/refresh:
    post:
      tags: [Auth]
      summary: Refresh access token by refresh cookie
      security:
        - refreshCookie: []
      responses:
        "200":
          description: Session refreshed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OKResponse'
        "401":
          description: Session expired
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/auth/logout:
    post:
      tags: [Auth]
      summary: Logout user
      security:
        - refreshCookie: []
      responses:
        "200":
          description: Logout successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OKResponse'
        "401":
          description: Session expired
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/me:
    get:
      tags: [User]
      summary: Get current user profile
      security:
        - accessCookie: []
      responses:
        "200":
          description: Current user profile
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MeResponse'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    patch:
      tags: [User]
      summary: Partially update current user profile
      security:
        - accessCookie: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PatchMeRequest'
      responses:
        "200":
          description: Updated user profile
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PatchMeResponse'
        "400":
          description: Validation failed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/places:
    get:
      tags: [Places]
      summary: List place cards
      responses:
        "200":
          description: Place cards list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PlaceCardsResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/places/{id}:
    get:
      tags: [Places]
      summary: Get place details by ID
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Place details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PlaceItemResponse'
        "404":
          description: Place not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/places/category/{category}:
    get:
      tags: [Places]
      summary: List place cards by category
      parameters:
        - in: path
          name: category
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Place cards by category
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PlaceCardsResponse'
        "404":
          description: Category not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/places/best:
    get:
      tags: [Places]
      summary: Get top places by like count
      responses:
        "200":
          description: Best places
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PlaceCardsResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/home:
    get:
      tags: [Places]
      summary: Get home payload
      responses:
        "200":
          description: Home page payload
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HomePayload'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
components:
  securitySchemes:
    accessCookie:
      type: apiKey
      in: cookie
      name: access_token
    refreshCookie:
      type: apiKey
      in: cookie
      name: refresh_token
  schemas:
    RegisterRequest:
      type: object
      required: [email, username, userSurname, password]
      properties:
        email:
          type: string
        username:
          type: string
        userSurname:
          type: string
        password:
          type: string
        birthday:
          type: string
          format: date
          nullable: true
        cityId:
          type: string
          nullable: true
    RegisterResponse:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
        username:
          type: string
        userSurname:
          type: string
        avatarUrl:
          type: string
          nullable: true
        createdAt:
          type: string
          format: date-time
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email:
          type: string
        password:
          type: string
    MessageResponse:
      type: object
      properties:
        message:
          type: string
    AccessTokenResponse:
      type: object
      properties:
        access_token:
          type: string
    OKResponse:
      type: object
      properties:
        ok:
          type: boolean
    LoginResponse:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
        username:
          type: string
    ErrorResponse:
      type: object
      properties:
        error:
          type: string
        details:
          type: object
          additionalProperties: true
    MeResponse:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
        username:
          type: string
        userSurname:
          type: string
        birthday:
          type: string
          format: date
          nullable: true
        avatarUrl:
          type: string
          nullable: true
        city:
          $ref: '#/components/schemas/UserCity'
        createdAt:
          type: string
          format: date-time
    PatchMeRequest:
      type: object
      properties:
        username:
          type: string
        userSurname:
          type: string
        birthday:
          type: string
          format: date
        cityId:
          type: string
        avatarUrl:
          type: string
    PatchMeResponse:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
        username:
          type: string
        userSurname:
          type: string
        birthday:
          type: string
          format: date
          nullable: true
        avatarUrl:
          type: string
          nullable: true
        updatedAt:
          type: string
          format: date-time
    UserCity:
      type: object
      nullable: true
      properties:
        id:
          type: string
        name:
          type: string
        countryName:
          type: string
        timezone:
          type: string
    PlaceCard:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        categories:
          type: array
          items:
            type: string
        like_count:
          type: integer
        location_description:
          type: string
        address:
          type: string
        image_url:
          type: string
    Place:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        categories:
          type: array
          items:
            type: string
        like_count:
          type: integer
        location_description:
          type: string
        full_description:
          type: string
        address:
          type: string
        image_url:
          type: string
        working_hours:
          type: string
        price_level:
          type: string
    PlaceCardsResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/PlaceCard'
    PlaceItemResponse:
      type: object
      properties:
        item:
          $ref: '#/components/schemas/Place'
    HomeTag:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        slug:
          type: string
    HomeNextSessionPlace:
      type: object
      properties:
        name:
          type: string
        addressLine:
          type: string
    HomeNextSession:
      type: object
      properties:
        startAt:
          type: string
          format: date-time
        place:
          $ref: '#/components/schemas/HomeNextSessionPlace'
    HomeFeaturedEvent:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        coverImageUrl:
          type: string
        tags:
          type: array
          items:
            $ref: '#/components/schemas/HomeTag'
        nextSession:
          $ref: '#/components/schemas/HomeNextSession'
    HomeCategory:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        slug:
          type: string
    HomeCollection:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        description:
          type: string
        imageUrl:
          type: string
    HomePayload:
      type: object
      properties:
        featuredEvents:
          type: array
          items:
            $ref: '#/components/schemas/HomeFeaturedEvent'
        categories:
          type: array
          items:
            $ref: '#/components/schemas/HomeCategory'
        collections:
          type: array
          items:
            $ref: '#/components/schemas/HomeCollection'
`

func OpenAPIYAMLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(openAPISpecYAML))
}

func SwaggerUIHandler(specPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>CityHawk Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: %q,
      dom_id: '#swagger-ui'
    });
  </script>
</body>
</html>`, specPath)
	}
}
