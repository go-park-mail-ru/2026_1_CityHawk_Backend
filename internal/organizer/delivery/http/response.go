package http

type applicationCreateResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type applicationResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	City          string `json:"city"`
	ProjectName   string `json:"projectName"`
	Categories    string `json:"categories"`
	Links         string `json:"links"`
	About         string `json:"about"`
	ReviewComment string `json:"reviewComment"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type applicationStatusResponse struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	ReviewComment string `json:"reviewComment"`
	UpdatedAt     string `json:"updatedAt"`
}
