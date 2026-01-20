package common

// ReviewStatus represents the review status of user-uploaded content
type ReviewStatus int

const (
	// ReviewStatusPending indicates the content is awaiting review
	ReviewStatusPending ReviewStatus = 0
	// ReviewStatusApproved indicates the content has been approved
	ReviewStatusApproved ReviewStatus = 1
	// ReviewStatusRejected indicates the content has been rejected
	ReviewStatusRejected ReviewStatus = 2
)
