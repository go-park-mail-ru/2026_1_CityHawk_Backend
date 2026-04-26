package model

import "time"

type Category string

const (
	CategoryBug              Category = "bug"
	CategorySuggestion       Category = "suggestion"
	CategoryProductComplaint Category = "product_complaint"
	CategoryOther            Category = "other"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryBug, CategorySuggestion, CategoryProductComplaint, CategoryOther:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusClosed     Status = "closed"
)

func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusClosed:
		return true
	default:
		return false
	}
}

type Ticket struct {
	ID        string
	UserID    string
	Category  Category
	Status    Status
	Title     string
	Message   string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
}

type MessageAuthorRole string

const (
	MessageAuthorRoleUser    MessageAuthorRole = "user"
	MessageAuthorRoleSupport MessageAuthorRole = "support"
	MessageAuthorRoleAdmin   MessageAuthorRole = "admin"
)

func (r MessageAuthorRole) Valid() bool {
	switch r {
	case MessageAuthorRoleUser, MessageAuthorRoleSupport, MessageAuthorRoleAdmin:
		return true
	default:
		return false
	}
}

type Message struct {
	ID           string
	TicketID     string
	AuthorUserID string
	AuthorRole   MessageAuthorRole
	Body         string
	CreatedAt    time.Time
}

type Filter struct {
	Status   Status
	Category Category
	Limit    int
	Offset   int
}

type StatsFilter struct {
	From *time.Time
	To   *time.Time
}

type Stats struct {
	Total           int
	ByStatus        map[Status]int
	ByCategory      map[Category]int
	OpenTotal       int
	InProgressTotal int
	ClosedTotal     int
}
