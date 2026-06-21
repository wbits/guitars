package guitaranalysis

// Job is an async photo analysis work item (one cover photo per guitar).
type Job struct {
	GuitarID string `json:"guitarId"`
	OwnerID  string `json:"ownerId"`
	Force    bool   `json:"force"`
}
