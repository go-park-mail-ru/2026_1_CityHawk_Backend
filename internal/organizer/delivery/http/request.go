package http

type createApplicationRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	City        string `json:"city"`
	ProjectName string `json:"projectName"`
	Categories  string `json:"categories"`
	Links       string `json:"links"`
	About       string `json:"about"`
	Consent     bool   `json:"consent"`
}

type updateApplicationStatusRequest struct {
	Status        string `json:"status"`
	ReviewComment string `json:"reviewComment"`
}
