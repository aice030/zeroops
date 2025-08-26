package models

// MockDataSource Mock数据源
type MockDataSource struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Data    map[string]any `json:"data"`
	Enabled bool           `json:"enabled"`
}
