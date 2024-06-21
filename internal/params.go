package internal

type DemoParams struct {
	Name  string `json:"name" form:"name"`
	Pause string `json:"pause" form:"pause"`
}
