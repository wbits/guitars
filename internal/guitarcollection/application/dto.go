package application

// GuitarInput is the data the application layer accepts when adding or
// updating a guitar. Keeping it free of HTTP/JSON concerns lets the same
// service be driven from a CLI, a test harness, or a future gRPC layer.
type GuitarInput struct {
	SerialNumber      string
	Color             string
	Country           string
	Factory           string
	Pictures          []string
	CoverPictureIndex int
	Description       string
	Brand             string
	TypeName          string
	BuildYear         int
	PriceAmount       int64
	PriceCurrency     string
}
