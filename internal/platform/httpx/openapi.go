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
  - name: Events
  - name: Search
  - name: Tags
  - name: Collections
  - name: Organizer
  - name: Support
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
        "302":
          description: Redirect to frontend home page
          headers:
            Location:
              schema:
                type: string
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
        "302":
          description: Redirect to frontend home page
          headers:
            Location:
              schema:
                type: string
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
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/PatchMeMultipartRequest'
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
  /api/events:
    get:
      tags: [Events]
      summary: List events
      parameters:
        - in: query
          name: query
          schema:
            type: string
        - in: query
          name: categoryId
          schema:
            type: string
        - in: query
          name: tagId
          schema:
            type: string
        - in: query
          name: cityId
          schema:
            type: string
        - in: query
          name: dateFrom
          schema:
            type: string
        - in: query
          name: dateTo
          schema:
            type: string
        - in: query
          name: authorId
          schema:
            type: string
        - in: query
          name: sort
          schema:
            type: string
        - in: query
          name: limit
          schema:
            type: integer
        - in: query
          name: offset
          schema:
            type: integer
      responses:
        "200":
          description: Event list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EventListResponse'
        "400":
          description: Validation failed
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
    post:
      tags: [Events]
      summary: Create event
      security:
        - accessCookie: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateEventRequest'
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/CreateEventMultipartRequest'
      responses:
        "201":
          description: Event created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EventIDResponse'
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
  /api/events/{id}:
    get:
      tags: [Events]
      summary: Get event details by ID
      parameters:
        - in: path
          name: id
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Event details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EventDetails'
        "404":
          description: Event not found
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
    patch:
      tags: [Events]
      summary: Update event
      security:
        - accessCookie: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/PatchEventRequest'
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/PatchEventMultipartRequest'
      responses:
        "200":
          description: Event updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/EventDetails'
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
        "403":
          description: Forbidden
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Event not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    delete:
      tags: [Events]
      summary: Delete event
      security:
        - accessCookie: []
      responses:
        "200":
          description: Event deleted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OKResponse'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "403":
          description: Forbidden
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Event not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/home:
    get:
      tags: [Events]
      summary: Get home payload
      parameters:
        - in: query
          name: city
          schema:
            type: string
          description: City id or city name
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
  /api/categories:
    get:
      tags: [Events]
      summary: Get categories
      responses:
        "200":
          description: Categories list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CategoriesResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/tags:
    get:
      tags: [Tags]
      summary: Get tags
      responses:
        "200":
          description: Tags list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TagsResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/cities:
    get:
      tags: [Cities]
      summary: Get cities
      responses:
        "200":
          description: Cities list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CitiesResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/collections:
    get:
      tags: [Collections]
      summary: Get public collections
      responses:
        "200":
          description: Collections list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CollectionsResponse'
        "405":
          description: Method not allowed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/collections/{collectionId}:
    get:
      tags: [Collections]
      summary: Get collection details
      parameters:
        - in: path
          name: collectionId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Collection details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CollectionDetails'
        "404":
          description: Collection not found
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
  /api/search:
    get:
      tags: [Search]
      summary: Get search suggestions
      parameters:
        - in: query
          name: query
          required: true
          schema:
            type: string
        - in: query
          name: limit
          required: false
          schema:
            type: integer
            minimum: 5
            maximum: 10
      responses:
        "200":
          description: Search suggestions
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SearchSuggestionsResponse'
        "400":
          description: Validation failed
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
  /api/organizer/applications:
    post:
      tags: [Organizer]
      summary: Create organizer application
      security:
        - accessCookie: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrganizerApplicationRequest'
      responses:
        "201":
          description: Organizer application created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrganizerApplicationCreateResponse'
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
        "403":
          description: CSRF token mismatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "409":
          description: Active application already exists
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/organizer/applications/me:
    get:
      tags: [Organizer]
      summary: Get current user's organizer application
      security:
        - accessCookie: []
      responses:
        "200":
          description: Organizer application
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrganizerApplication'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Organizer application not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/admin/organizer/applications/{applicationId}:
    patch:
      tags: [Organizer]
      summary: Update organizer application status
      security:
        - accessCookie: []
      parameters:
        - in: path
          name: applicationId
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateOrganizerApplicationStatusRequest'
      responses:
        "200":
          description: Organizer application status updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrganizerApplicationStatusResponse'
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
        "403":
          description: Forbidden or CSRF token mismatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Organizer application not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/support/tickets:
    get:
      tags: [Support]
      summary: List current user's support tickets
      security:
        - accessCookie: []
      parameters:
        - in: query
          name: status
          schema:
            type: string
            enum: [open, in_progress, closed]
        - in: query
          name: category
          schema:
            type: string
            enum: [bug, suggestion, product_complaint, other]
        - in: query
          name: limit
          schema:
            type: integer
            default: 20
            maximum: 100
        - in: query
          name: offset
          schema:
            type: integer
            default: 0
      responses:
        "200":
          description: Support tickets
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicketListResponse'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    post:
      tags: [Support]
      summary: Create support ticket
      security:
        - accessCookie: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateSupportTicketRequest'
      responses:
        "201":
          description: Created support ticket
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicket'
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
        "403":
          description: CSRF token mismatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/support/tickets/{ticketId}:
    get:
      tags: [Support]
      summary: Get current user's support ticket
      security:
        - accessCookie: []
      parameters:
        - in: path
          name: ticketId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Support ticket
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicket'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Support ticket not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    patch:
      tags: [Support]
      summary: Update current user's support ticket
      security:
        - accessCookie: []
      parameters:
        - in: path
          name: ticketId
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateSupportTicketRequest'
      responses:
        "200":
          description: Updated support ticket
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicket'
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
        "403":
          description: CSRF token mismatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "403":
          description: Forbidden
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Support ticket not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "409":
          description: Support ticket is closed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/support/tickets/{ticketId}/status:
    patch:
      tags: [Support]
      summary: Update support ticket status
      security:
        - accessCookie: []
      parameters:
        - in: path
          name: ticketId
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateSupportTicketStatusRequest'
      responses:
        "200":
          description: Updated support ticket
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicket'
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
        "403":
          description: Forbidden or CSRF token mismatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Support ticket not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/support/tickets/{ticketId}/messages:
    get:
      tags: [Support]
      summary: List support ticket messages
      security:
        - accessCookie: []
      parameters:
        - in: path
          name: ticketId
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Support ticket messages
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicketMessageListResponse'
        "401":
          description: Unauthorized
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Support ticket not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
    post:
      tags: [Support]
      summary: Create support ticket message
      security:
        - accessCookie: []
      parameters:
        - in: path
          name: ticketId
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateSupportTicketMessageRequest'
      responses:
        "201":
          description: Created support ticket message
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicketMessage'
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
        "403":
          description: CSRF token mismatch
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
        "404":
          description: Support ticket not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
  /api/support/stats:
    get:
      tags: [Support]
      summary: Get support ticket statistics
      security:
        - accessCookie: []
      parameters:
        - in: query
          name: from
          schema:
            type: string
            format: date-time
        - in: query
          name: to
          schema:
            type: string
            format: date-time
      responses:
        "200":
          description: Support ticket statistics
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/SupportTicketStats'
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
        "403":
          description: Forbidden
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
      required: [email, password]
      properties:
        email:
          type: string
        password:
          type: string
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
        role:
          type: string
          enum: [user, organizer, admin]
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
        email:
          type: string
          format: email
        username:
          type: string
        userSurname:
          type: string
        role:
          type: string
          enum: [user, organizer, admin]
        birthday:
          type: string
          format: date
        cityId:
          type: string
        avatarUrl:
          type: string
    PatchMeMultipartRequest:
      type: object
      properties:
        email:
          type: string
          format: email
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
        avatar:
          type: string
          format: binary
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
    EventTaxonomyItem:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        slug:
          type: string
    EventCardNextSessionPlace:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        addressLine:
          type: string
        latitude:
          type: number
        longitude:
          type: number
        city:
          $ref: '#/components/schemas/EventSessionPlaceCity'
    EventCardNextSession:
      type: object
      properties:
        startAt:
          type: string
          format: date-time
        place:
          $ref: '#/components/schemas/EventCardNextSessionPlace'
    EventCard:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        shortDescription:
          type: string
        coverImageUrl:
          type: string
        tags:
          type: array
          items:
            $ref: '#/components/schemas/EventTaxonomyItem'
        nextSession:
          $ref: '#/components/schemas/EventCardNextSession'
    EventListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/EventCard'
        total:
          type: integer
        limit:
          type: integer
        offset:
          type: integer
    EventAuthor:
      type: object
      properties:
        id:
          type: string
        username:
          type: string
        avatarUrl:
          type: string
          nullable: true
    EventImage:
      type: object
      properties:
        id:
          type: string
        imageUrl:
          type: string
    EventSessionPlaceCity:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        countryName:
          type: string
        timezone:
          type: string
    EventSessionPlace:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        addressLine:
          type: string
        latitude:
          type: number
        longitude:
          type: number
        city:
          $ref: '#/components/schemas/EventSessionPlaceCity'
    EventSession:
      type: object
      properties:
        id:
          type: string
        startAt:
          type: string
          format: date-time
        endAt:
          type: string
          format: date-time
        price:
          type: integer
        placeName:
          type: string
        place:
          $ref: '#/components/schemas/EventSessionPlace'
    EventDetails:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        shortDescription:
          type: string
        fullDescription:
          type: string
        ageLimit:
          type: integer
        sourceUrl:
          type: string
          nullable: true
        author:
          $ref: '#/components/schemas/EventAuthor'
        placeName:
          type: string
        place:
          $ref: '#/components/schemas/EventSessionPlace'
          nullable: true
        categories:
          type: array
          items:
            $ref: '#/components/schemas/EventTaxonomyItem'
        tags:
          type: array
          items:
            $ref: '#/components/schemas/EventTaxonomyItem'
        images:
          type: array
          items:
            $ref: '#/components/schemas/EventImage'
        sessions:
          type: array
          items:
            $ref: '#/components/schemas/EventSession'
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
        isFavorite:
          type: boolean
        isOwner:
          type: boolean
    EventSessionInput:
      type: object
      required: [placeId, startAt, endAt, price]
      properties:
        placeId:
          type: string
        placeName:
          type: string
        startAt:
          type: string
          format: date-time
        endAt:
          type: string
          format: date-time
        price:
          type: integer
    CreateEventRequest:
      type: object
      required: [title, shortDescription, fullDescription, categoryIds]
      properties:
        title:
          type: string
        shortDescription:
          type: string
        fullDescription:
          type: string
        ageLimit:
          type: integer
        sourceUrl:
          type: string
          nullable: true
        categoryIds:
          type: array
          items:
            type: string
        tagIds:
          type: array
          items:
            type: string
        imageUrls:
          type: array
          items:
            type: string
        placeId:
          type: string
          nullable: true
        sessions:
          type: array
          items:
            $ref: '#/components/schemas/EventSessionInput'
    PatchEventRequest:
      type: object
      properties:
        title:
          type: string
        shortDescription:
          type: string
        fullDescription:
          type: string
        ageLimit:
          type: integer
        sourceUrl:
          type: string
          nullable: true
        categoryIds:
          type: array
          items:
            type: string
        tagIds:
          type: array
          items:
            type: string
        imageUrls:
          type: array
          items:
            type: string
        placeId:
          type: string
          nullable: true
        sessions:
          type: array
          items:
            $ref: '#/components/schemas/EventSessionInput'
    CreateEventMultipartRequest:
      type: object
      required: [title, shortDescription, fullDescription, categoryIds]
      properties:
        title:
          type: string
        shortDescription:
          type: string
        fullDescription:
          type: string
        ageLimit:
          type: integer
        sourceUrl:
          type: string
          nullable: true
        categoryIds:
          type: string
          description: JSON array of strings
        tagIds:
          type: string
          description: JSON array of strings
        imageUrls:
          type: string
          description: JSON array of strings
        placeId:
          type: string
          nullable: true
        sessions:
          type: string
          description: JSON array of session objects
        images:
          type: array
          items:
            type: string
            format: binary
    PatchEventMultipartRequest:
      type: object
      properties:
        title:
          type: string
        shortDescription:
          type: string
        fullDescription:
          type: string
        ageLimit:
          type: integer
        sourceUrl:
          type: string
          nullable: true
        categoryIds:
          type: string
          description: JSON array of strings
        tagIds:
          type: string
          description: JSON array of strings
        imageUrls:
          type: string
          description: JSON array of strings
        placeId:
          type: string
          nullable: true
        sessions:
          type: string
          description: JSON array of session objects
        images:
          type: array
          items:
            type: string
            format: binary
    EventIDResponse:
      type: object
      properties:
        id:
          type: string
    CategoriesResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/EventTaxonomyItem'
    TagsResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/EventTaxonomyItem'
    CityItem:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        countryName:
          type: string
        timezone:
          type: string
    CitiesResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/CityItem'
    CollectionCard:
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
        isPublic:
          type: boolean
    CollectionsResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/CollectionCard'
    CollectionDetails:
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
        isPublic:
          type: boolean
        events:
          type: array
          items:
            $ref: '#/components/schemas/EventCard'
    SearchSuggestionsResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/SearchSuggestion'
    SearchSuggestion:
      type: object
      properties:
        id:
          type: string
        type:
          type: string
          enum: [user, event, category, tag]
        title:
          type: string
        label:
          type: string
        avatarUrl:
          type: string
          nullable: true
        isFollowing:
          type: boolean
    CreateOrganizerApplicationRequest:
      type: object
      required: [name, email, phone, city, projectName, categories, about, consent]
      properties:
        name:
          type: string
        email:
          type: string
          format: email
        phone:
          type: string
        city:
          type: string
        projectName:
          type: string
        categories:
          type: string
        links:
          type: string
        about:
          type: string
        consent:
          type: boolean
    UpdateOrganizerApplicationStatusRequest:
      type: object
      required: [status]
      properties:
        status:
          type: string
          enum: [pending, needs_info, approved, rejected]
        reviewComment:
          type: string
    OrganizerApplicationCreateResponse:
      type: object
      properties:
        id:
          type: string
        status:
          type: string
          enum: [pending, needs_info, approved, rejected]
        createdAt:
          type: string
          format: date-time
    OrganizerApplication:
      type: object
      properties:
        id:
          type: string
        status:
          type: string
          enum: [pending, needs_info, approved, rejected]
        name:
          type: string
        email:
          type: string
        phone:
          type: string
        city:
          type: string
        projectName:
          type: string
        categories:
          type: string
        links:
          type: string
        about:
          type: string
        reviewComment:
          type: string
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
    OrganizerApplicationStatusResponse:
      type: object
      properties:
        id:
          type: string
        status:
          type: string
          enum: [pending, needs_info, approved, rejected]
        reviewComment:
          type: string
        updatedAt:
          type: string
          format: date-time
    SupportTicket:
      type: object
      properties:
        id:
          type: string
        category:
          type: string
          enum: [bug, suggestion, product_complaint, other]
        status:
          type: string
          enum: [open, in_progress, closed]
        title:
          type: string
        message:
          type: string
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time
        closedAt:
          type: string
          format: date-time
          nullable: true
    SupportTicketListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/SupportTicket'
        limit:
          type: integer
        offset:
          type: integer
    CreateSupportTicketRequest:
      type: object
      required: [category, title, message]
      properties:
        category:
          type: string
          enum: [bug, suggestion, product_complaint, other]
        title:
          type: string
        message:
          type: string
    UpdateSupportTicketRequest:
      type: object
      properties:
        category:
          type: string
          enum: [bug, suggestion, product_complaint, other]
        title:
          type: string
        message:
          type: string
    UpdateSupportTicketStatusRequest:
      type: object
      required: [status]
      properties:
        status:
          type: string
          enum: [open, in_progress, closed]
    SupportTicketMessage:
      type: object
      properties:
        id:
          type: string
        ticketId:
          type: string
        authorUserId:
          type: string
        authorRole:
          type: string
          enum: [user, support, admin]
        body:
          type: string
        createdAt:
          type: string
          format: date-time
    SupportTicketMessageListResponse:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/SupportTicketMessage'
    CreateSupportTicketMessageRequest:
      type: object
      required: [body]
      properties:
        body:
          type: string
    SupportTicketStats:
      type: object
      properties:
        total:
          type: integer
        byStatus:
          type: object
          additionalProperties:
            type: integer
        byCategory:
          type: object
          additionalProperties:
            type: integer
        openTotal:
          type: integer
        inProgressTotal:
          type: integer
        closedTotal:
          type: integer
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
