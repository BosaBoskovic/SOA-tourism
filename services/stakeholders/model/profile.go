package model

type Profile struct {
	Username  string
	FirstName string
	LastName  string
	ImageURL  string
	Bio       string
	Motto     string
}

type ProfileResponse struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	ImageURL  string `json:"imageURL"`
	Bio       string `json:"bio"`
	Motto     string `json:"motto"`
}

type UpdateProfileRequest struct {
	FirstName string `json:"firstName" binding:"omitempty,max=50"`
	LastName  string `json:"lastName"  binding:"omitempty,max=50"`
	ImageURL  string `json:"imageURL"  binding:"omitempty,url"`
	Bio       string `json:"bio"       binding:"omitempty,max=1000"`
	Motto     string `json:"motto"     binding:"omitempty,max=200"`
}
